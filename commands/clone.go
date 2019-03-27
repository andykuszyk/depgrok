package commands

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

// Clone will clone all repos from a GitHub organisation, with minimal depth, so that
// their contents can be traversed by the `search` command
func Clone(c *cli.Context) {
	org := c.String("org")
	token := c.String("token")
	dir := c.String("dir")
	if org == "" || token == "" || dir == "" {
		log.Fatal("--org, --token and --dir are required flags for the `clone` command")
	}

	httpClient := http.Client{}
	page := 1
	repos := []map[string]interface{}{}
	for {
		request, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/orgs/%s/repos?page=%d", org, page), nil)
		request.Header.Add("Authorization", "token "+token)
		res, err := httpClient.Do(request)
		if err != nil {
			log.Fatal("Error making request to GitHub:", err)
		}
		data, _ := ioutil.ReadAll(res.Body)
		var responseRepos []map[string]interface{}
		json.Unmarshal(data, &responseRepos)
		if len(responseRepos) == 0 {
			break
		}
		for _, repo := range responseRepos {
			repos = append(repos, repo)
		}
		page += 1
	}

	fmt.Printf("Found %d repos, cloning them into `dir`...\n", len(repos))
	for _, repo := range repos {
		fmt.Printf("Cloning %v...\n", repo["name"])
		gitClone := exec.Command("git", "clone", repo["ssh_url"].(string), "--depth", "1")
		gitClone.Dir = dir
		err := gitClone.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

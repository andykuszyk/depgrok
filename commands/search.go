package commands

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type dependant struct {
	name   string
	parent *dependant
	repos  map[string]bool
	level  int
}

func (d *dependant) addRepo(repo string) {
	if d.repos == nil {
		d.repos = make(map[string]bool)
	}
	d.repos[repo] = true
}

// searchChildren is a recursive function that searches a directory tree for references to dependants.
func searchChildren(repo string, parent string, dependants []*dependant, level int) []*dependant {
	parentInfo, _ := os.Stat(parent)
	if parentInfo.IsDir() {
		children, _ := ioutil.ReadDir(parent)
		for _, child := range children {
			if strings.HasPrefix(child.Name(), ".") || child.Name() == "bin" {
				continue
			}
			newRepo := repo
			if newRepo == "" {
				newRepo = child.Name()
			}
			dependants = searchChildren(newRepo, filepath.Join(parent, child.Name()), dependants, level)
		}
	} else {
		bytes, err := ioutil.ReadFile(parent)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", parent, err)
		}
		text := string(bytes)
		for _, dep := range dependants {
			if dep.level != level {
				continue
			}
			if strings.Contains(text, dep.name) {
				dep.addRepo(repo)
				dependants = append(dependants, &dependant{name: parentInfo.Name(), parent: dep, level: level + 1})
			}
		}
	}
	return dependants
}

type dependancyDiagram struct {
	text          string
	dependantName string
	repoName      string
}

func (d *dependant) buildDependencyDiagram(repo string) dependancyDiagram {
	text := fmt.Sprintf("%s -> %s", repo, d.name)
	depName := d.name
	parent := d.parent
	for parent != nil {
		depName = parent.name
		text = fmt.Sprintf("%s -> %s", text, parent.name)
		parent = parent.parent
	}
	return dependancyDiagram{
		text:          text,
		dependantName: depName,
		repoName:      repo,
	}
}

func Search(c *cli.Context) {
	deps := c.String("deps")
	dir := c.String("dir")
	depth := c.Int("depth")
	if dir == "" || deps == "" {
		log.Fatal("--deps and --dir are required flags")
	}

	dependants := []*dependant{}
	for _, item := range strings.Fields(deps) {
		dependants = append(dependants, &dependant{name: item, level: 0})
	}

	for i := 0; i < depth; i++ {
		dependants = searchChildren("", dir, dependants, i)
	}

	diagramsByDepRepo := make(map[string]map[string]dependancyDiagram)
	for _, dep := range dependants {
		for repo, _ := range dep.repos {
			diagram := dep.buildDependencyDiagram(repo)
			if diagramsByDepRepo[diagram.dependantName] == nil {
				diagramsByDepRepo[diagram.dependantName] = make(map[string]dependancyDiagram)
			}
			diagramsByDepRepo[diagram.dependantName][diagram.repoName] = diagram
		}
	}

	depKeys := []string{}
	for depKey, _ := range diagramsByDepRepo {
		depKeys = append(depKeys, depKey)
	}
	sort.Strings(depKeys)
	for _, depKey := range depKeys {
		fmt.Printf("# %s:\n", depKey)
		diagramsByRepo := diagramsByDepRepo[depKey]
		repos := []string{}
		for repo, _ := range diagramsByRepo {
			repos = append(repos, repo)
		}
		sort.Strings(repos)
		for _, repo := range repos {
			diagram := diagramsByRepo[repo]
			fmt.Println(diagram.text)
		}
		fmt.Println("")
	}
}

package commands

import (
    "github.com/urfave/cli"
    "fmt"
    "strings"
    "log"
    "io/ioutil"
    "path/filepath"
)

type dependant struct {
    name string
    parent *dependant
    repos []string
}

func (d *dependant) addRepo(repo string) {
    d.repos = append(d.repos, repo)
}

// searchChildren is a recursive function that searches a directory tree
// for references to dependants.
//
// repo is the root directory of a search. If nil is provided, the path of any
// sub-directories that are searched is used.
//
// parent is the name of the current directory (just the directory name) being searched.
func searchChildren(repo string, parent string, dependants []dependant) {
    children, err := ioutil.ReadDir(parent)
    if err != nil {
        log.Fatalf("Error reading %s: %v", parent, err)
    }

    if len(children) > 0 {
        for _, child := range children {
            if strings.HasPrefix(child.Name(), ".") {
                continue
            }
            newRepo := repo
            if newRepo == "" {
                newRepo = child.Name()
            }
            searchChildren(newRepo, filepath.Join(parent, child.Name()), dependants)
        }
    }

    bytes, err := ioutil.ReadFile(parent)
    text := string(bytes)
    for _, dep := range dependants {
        if strings.Contains(text, dep.name) {
            dep.addRepo(repo)
        }
    }
}

func Search(c *cli.Context) {
    deps := c.String("deps")
    dir := c.String("dir")
    if dir == "" || deps == "" {
        log.Fatal("--deps and --dir are required flags")
    }

    dependants := []dependant{}
    for _, item := range strings.Fields(deps) {
        dependants = append(dependants, dependant { name: item })
    }

    searchChildren("", dir, dependants)
    for _, dep := range dependants {
        for _, repo := range dep.repos {
            fmt.Printf("%s -> %s\n", repo, dep)
        }
    }
}

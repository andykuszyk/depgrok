package commands

import (
    "github.com/urfave/cli"
    "fmt"
    "strings"
    "log"
    "io/ioutil"
)

type dependant struct {
    name string
    parent *dep
    repos []string
}

func searchChildren(parent string, dependants *[]dependant) {
    children, err := ioutil.ReadDir(parent)
    if err != nil {
        log.Fatalf("Error reading %s: %v", parent, err)
    }

    if len(children) > 0 {
        for child, i := range children {
            searchChildren(parent, dependants)
        }
    }

    // TODO: Need to perform search here
}

func Search(c *cli.Context) {
    deps = c.String("deps")
    dir = c.String("dir")
    if dir == "" || deps == "" {
        log.Fatal("--deps and --dir are required flags")
    }

    dependants := []dependant{}
    for _, item := range strings.Fields(deps) {
        dependants = append(dependants, dependant { name: item })
    }

    searchChildren(dir, &dependants)
}

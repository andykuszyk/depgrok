package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/andykuszyk/depgrok/deps"
)

// Returns true if the given parent is a valid candidate for a search,
// otherwise false is returned (e.g. in the case of a filename beginning
// with ".".
func isValidParent(parent string) bool {
	return !(strings.HasPrefix(parent, ".") || parent == "bin")
}

// Recursively searches a file tree, amending and augmenting dependencies (at the given
// level) as matches are discovered.
func searchChildren(repo string, parent string, dependencies *deps.Dependencies, level int) {
	parentInfo, err := os.Stat(parent)
	if err != nil {
		log.Fatalf("An error occured calling os.Stat(%s): %v", parent, err)
	}

	// First, check if the parent location is a directory. If it is, traverse its children, if not
	// interrogate its contents.
	if parentInfo.IsDir() {
		children, err := ioutil.ReadDir(parent)
		if err != nil {
			log.Fatalf("An error occured calling ioutil.ReadDir(%s): %v", parent, err)
		}

		// Traverse the children of the parent, making a recursive call to searchChildren
		// if its a valid child.
		for _, child := range children {
			if !isValidParent(child.Name()) {
				continue
			}
			// If this is the first traversal of the root directory, no repo name will have been
			// provided, so extract this from the child's name.
			newRepo := repo
			if newRepo == "" {
				newRepo = child.Name()
			}
			searchChildren(newRepo, filepath.Join(parent, child.Name()), dependencies, level)
		}
	} else {
		// Interrogate the file - read its contents out as a string.
		bytes, err := ioutil.ReadFile(parent)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", parent, err)
		}
		text := string(bytes)

		// Now, iterate though each of the dependencies at the current level (in order avoid
		// worrying about new dependencies of a higher level that have been collected on this pass)
		// and check for a reference within the file.
		for _, dep := range dependencies.Slice() {
			if dep.Level != level {
				continue
			}
			if dep.Matches(text) {
				dep.AddRepo(repo)
				parentDependency := deps.Dependency{
					Name: parentInfo.Name(),
					Parent: dep,
					Level: level + 1,
				}
				if !dependencies.Contains(parentDependency) {
					dependencies.Add(&parentDependency)
				}
			}
		}
	}
}

func Search(c *cli.Context) {
	depsArg := c.String("deps")
	dir := c.String("dir")
	depth := c.Int("depth")
	if dir == "" || depsArg == "" {
		log.Fatal("--deps and --dir are required flags")
	}

	// Construct list of dependencies and collect repo relationships
	// by searching children.
	dependencies := deps.BuildDependencies(strings.Fields(depsArg))
	for i := 0; i < depth; i++ {
		searchChildren("", dir, dependencies, i)
	}

	// Print out diagrams to screen in a reasonable order.
	for _, diagram := range dependencies.BuildDiagrams() {
		fmt.Println(diagram.Text)
	}
}

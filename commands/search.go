package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"regexp"

	"github.com/andykuszyk/depgrok/deps"
	"github.com/urfave/cli"
)

// Returns true if the given parent is a valid candidate for a search,
// otherwise false is returned (e.g. in the case of a filename beginning
// with ".".
func isValidParent(parent string) bool {
	return !(strings.HasPrefix(parent, ".") || parent == "bin" || parent == "obj")
}

var secondsBetweenLoadingChars = 1.0
func getNextLoadingCharFunc() func() string {
	char := "|"
	lastCharTime := time.Now()
	return func() string {
		if time.Now().Sub(lastCharTime).Seconds() < secondsBetweenLoadingChars {
			return char
		}
		lastCharTime = time.Now()
		switch char {
		case "|":
			char = "/"
			return char
		case "/":
			char = "-"
			return char
		case "-":
			char = "\\"
			return char
		case "\\":
			char = "|"
			return char
		}
		return char
	}
}

var getNextLoadingChar = getNextLoadingCharFunc()

// Gets a FileInfo for the given path, returning it if successful, or
// logging a fatal error if not.
func getFileInfo(path string) os.FileInfo {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("An error occured calling os.Stat(%s): %v", path, err)
	}
	return fileInfo
}

// Controls whether or not searches of the file system will be parallelised.
// By default, this is true, although it can be set to false in tests to allow
// go routines to be debugged.
var paralleliseSearches = true

func matchGlob(path string, globs []string) []string {
	globMatches := []string{}
	for _, glob := range globs {
		matches, err := filepath.Glob(filepath.Join(path, glob))
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occured handling the glob (%s): %v", glob, err)
			continue
		}
		globMatches = append(globMatches, matches...)
	}
	return globMatches
}

// Iterates over the children of a given parent path, calling searchChildren as a 
// go routine on each.
func traverseChildren(
	parent string,
	dependencies *deps.Dependencies,
	level int,
	wg *sync.WaitGroup,
	repo string,
	exclude []string,
	include []string,
	repos chan repoCount) {
	children, err := ioutil.ReadDir(parent)
	if err != nil {
		log.Fatalf("An error occured calling ioutil.ReadDir(%s): %v", parent, err)
	}

	// See if any of the children in this parent match the exclusions provided, by first
	// building up a list of files to exclude.
	excludeMatches := matchGlob(parent, exclude)

	// Traverse the children of the parent, making a recursive call to searchChildren
	// if its a valid child.
	for _, child := range children {
		// First, check that the file name is a valid one for searching.
		if !isValidParent(child.Name()) {
			continue
		}

		// Then, check that the file is not supposed to be excluded by one of the provided
		// globs.
		excludeFile := false
		for _, excludeMatch := range excludeMatches {
			if filepath.Join(parent, child.Name()) == excludeMatch {
				excludeFile = true
				break
			}
		}
		if excludeFile {
			continue
		}

		// If this is the first traversal of the root directory, no repo name will have been
		// provided, so extract this from the child's name.
		newRepo := repo
		if newRepo == "" {
			newRepo = child.Name()
		}
		if paralleliseSearches {
			go searchChildren(newRepo, filepath.Join(parent, child.Name()), dependencies, level, wg, exclude, include, repos)
		} else {
			searchChildren(newRepo, filepath.Join(parent, child.Name()), dependencies, level, wg, exclude, include, repos)
		}
	}
}

// Searches the file at the path parent for references to the given dependencies,
// updating or augmenting the dependencies list as and when matches are found.
func searchFile(parent string, repo string, dependencies *deps.Dependencies, parentInfo os.FileInfo, level int) {
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
		if stripExtension(parentInfo.Name()) != dep.Name && dep.Matches(text) {
			dep.AddRepo(repo)
			parentDependency := deps.Dependency{
				Name:   stripExtension(parentInfo.Name()),
				Parent: dep,
				Level:  level + 1,
			}
			if !dependencies.Contains(parentDependency) {
				dependencies.Add(&parentDependency)
			}
		}
	}
}

// Recursively searches a file tree, amending and augmenting dependencies (at the given
// level) as matches are discovered.
func searchChildren(
	repo string,
	parent string,
	dependencies *deps.Dependencies,
	level int,
	wg *sync.WaitGroup,
	exclude []string,
	include []string,
	repos chan repoCount) {
	// Ensure that we add a counter to the waitgroup for this function call,
	// and also wait on the "semaphore" channel to ensure too many parallel
	// executions of this function are not taking place.
	wg.Add(1)
	defer wg.Done()
	sem <- 1
	releaseSem := func() {
		<-sem
	}
	defer releaseSem()

	// First, check if the parent location is a directory. If it is, traverse its children, if not
	// interrogate its contents.
	parentInfo := getFileInfo(parent)
	if parentInfo.IsDir() {
		repos <- repoCount{Level: level, Count: 1, Path: parent}
		traverseChildren(parent, dependencies, level, wg, repo, exclude, include, repos)
	} else {
		// Also, check that the file is supposed to be included.
		if len(include) > 0 {
			includeMatches := matchGlob(filepath.Dir(parent), include)
			for _, includeMatch := range includeMatches {
				if parent == includeMatch {
					repos <- repoCount{Level: level, Count: 1, Path: parent}
					searchFile(parent, repo, dependencies, parentInfo, level)
				}
			}
		} else {
			repos <- repoCount{Level: level, Count: 1, Path: parent}
			searchFile(parent, repo, dependencies, parentInfo, level)
		}
	}
}

// A closed function encapsulating the regexp to strip extensions from
// file names.
func stripExtensionFunc() func(string) string {
	stripExtensionRegexp, err := regexp.Compile("(.*)\\.[a-zA-Z]+$")
	if err != nil {
		log.Fatalf("Error parsing strip extension regex: %v", err)
	}
	return func(filename string) string {
		return stripExtensionRegexp.ReplaceAllString(filename, "$1")
	}
}

// Strips extensions from file names.
var stripExtension = stripExtensionFunc()

func logDuration(start time.Time, msg string) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "%s: %v", msg, time.Now().Sub(start).Seconds())
	fmt.Fprintln(os.Stderr, "")
}

var sem = make(chan int, runtime.NumCPU()*2)

type repoCount struct {
	Count int
	Level int
	Path string
}

func logRepos(repos chan repoCount, debug bool) {
	repoCountByLevel := map[int]int{}
	for {
		repoCount, _ := <-repos
		repoCountByLevel[repoCount.Level] += repoCount.Count
		if debug {
			fmt.Fprintln(os.Stderr, repoCount.Path)
		} else {
			fmt.Fprintf(os.Stderr, "\rSearching level %d files: %d", repoCount.Level, repoCountByLevel[repoCount.Level])
		}
	}
}

// The main function for the search command - setups up concurrency primitives and
// initiates a search of the required depth, collecting depdencies and formating
// them for output on the console.
func Search(c *cli.Context) {
	depsArg := c.String("deps")
	dir := c.String("dir")
	depth := c.Int("depth")
	if dir == "" || depsArg == "" {
		log.Fatal("--deps and --dir are required flags")
	}
	exclude := c.StringSlice("exclude")
	include := c.StringSlice("include")
	if len(exclude) > 0 && len(include) > 0 {
		log.Fatal("--exclude cannot be used in conjunction with --include")
	}
	debug := c.Bool("debug")

	// Record the time now and defer a timer until after execution is complete.
	defer logDuration(time.Now(), "Total time")

	// Construct list of dependencies and collect repo relationships
	// by searching children.
	wg := sync.WaitGroup{}
	dependencies := deps.BuildDependencies(strings.Fields(depsArg))
	repos := make(chan repoCount)
	go logRepos(repos, debug)
	start := time.Now()
	for i := 0; i < depth; i++ {
		searchChildren("", dir, dependencies, i, &wg, exclude, include, repos)
		wg.Wait()
	}
	logDuration(start, "SearchChildren")

	// Print out diagrams to screen in a reasonable order.
	fmt.Println("")
	start = time.Now()
	for _, diagram := range dependencies.BuildDiagrams() {
		fmt.Println(diagram.Text)
	}
	logDuration(start, "BuildDiagrams")
}

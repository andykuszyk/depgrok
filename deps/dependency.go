package deps

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"regexp"
)

// Represents a dependency that is being searched for, which might be related to
// one or more repos once the search is complete, or another dependency of a higher
// level.
//
// e.g. a reference to an object or entity X might be represented as a Dependency of Name
// X, which is related to a number of repositories, if it has a Level of 1. It might also
// be related to a repository via an intermediate dependency of level 1, if it has a deeper
// level (2, for example).
type Dependency struct {
	Name   string
	Parent *Dependency
	Repos  map[string]bool
	Level  int
	mutex sync.Mutex
	regexp *regexp.Regexp
}

// Adds a new repo to a Dependency's Repos map, setting its mapped value to true,
// which allows the Repos map to be used as a set.
func (d *Dependency) AddRepo(repo string) {
	if d.Repos == nil {
		d.Repos = make(map[string]bool)
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Repos[repo] = true
}

// Determines whether or not the Dependency is referenced in the given text.
func (d *Dependency) Matches(text string) bool {
	if d.regexp == nil {
		r, err := regexp.Compile(fmt.Sprintf("[^a-zA-Z]%s[^a-zA-Z]", d.Name))
		if err != nil {
			return false
		}
		d.regexp = r
	}
	return d.regexp.FindString(text) != ""
}

// Represents a diagram illustrating the relationship between a Dependency and
// a repo, along with additional information to help it be sorted into a convenient
// display.
type DependencyDiagram struct {
	Text           string
	DependencyName string
	RepoName       string
}

// Constructs a DependencyDiagram from the information stored in a Dependency,
// accounting for the information available via the ancestry of the Parent Dependency.
func (d *Dependency) DependencyDiagram(repo string) DependencyDiagram {
	text := fmt.Sprintf("%s -> %s", repo, d.Name)
	depName := d.Name
	parent := d.Parent
	for parent != nil {
		depName = parent.Name
		text = fmt.Sprintf("%s -> %s", text, parent.Name)
		parent = parent.Parent
	}
	return DependencyDiagram{
		Text:           text,
		DependencyName: depName,
		RepoName:       repo,
	}
}

// A sortable key value representing a `DependencyDiagram`.
type dependencyDiagramKey struct {
	Name string
	Repo string
	Level int
}

// Represents an array of `dependencyDiagramKey`s, which can be sorted using `sort.Sort()`.
type dependencyDiagramKeys []dependencyDiagramKey

// Provides a `Len` implementation for `sort.Interface`.
func (k dependencyDiagramKeys) Len() int {
	return len(k)
}

// Provides a `Less` implementation for `sort.Interface`, sorting keys
// repo, name and level.
func (k dependencyDiagramKeys) Less(i, j int) bool {
	elementI := k[i]
	elementJ := k[j]
	if elementI.Repo < elementJ.Repo {
		return true
	}
	if elementI.Repo > elementJ.Repo {
		return false
	}

	if elementI.Name < elementJ.Name {
		return true
	}
	if elementI.Name > elementJ.Name {
		return false
	}

	return elementI.Level < elementJ.Level
}

// Provides a `Swap` implementation for `sort.Interface`.
func (k dependencyDiagramKeys) Swap(i, j int) {
	elementI := k[i]
	elementJ := k[j]
	k[i] = elementJ
	k[j] = elementI
}

// Constructs a set of DependencyDiagram representations from the Dependencies collection
// and returns them grouped, sorted and de-duplicated, ready for display to the user.
func (d *Dependencies) BuildDiagrams() []DependencyDiagram {
	// First, make a list of DependencyDiagrams. This list will contain ordered diagrams,
	// arranged ready to display to the user.
	diagrams := []DependencyDiagram{}

	// Build a map of DependencyDiagrams, keyed by a dependencyDiagramKey.
	diagramsByKey := make(map[dependencyDiagramKey]DependencyDiagram)
	for _, dep := range d.dependencies {
		for repo, _ := range dep.Repos {
			diagramsByKey[dependencyDiagramKey{Name: dep.Name, Repo: repo, Level: dep.Level}] = dep.DependencyDiagram(repo)
		}
	}

	depKeys := dependencyDiagramKeys{}
	for depKey, _ := range diagramsByKey {
		depKeys = append(depKeys, depKey)
	}
	sort.Sort(&depKeys)
	sort.Sort(&depKeys)
	sort.Sort(&depKeys)
	for _, depKey := range depKeys {
		diagrams = append(diagrams, diagramsByKey[depKey])
	}

	return diagrams
}

// Represents a collection of Dependency types, encapsulating accessing, membership
// checking and additions to the collection.
type Dependencies struct {
	dependencies map[string]*Dependency
	membership   map[string]bool
	mutex        sync.Mutex
}

// Constructs a new Dependencies collection from the given list of dependency names.
func BuildDependencies(dependencies []string) *Dependencies {
	deps := Dependencies{}
	deps.membership = make(map[string]bool)
	deps.dependencies = make(map[string]*Dependency)
	for _, item := range dependencies {
		deps.Add(&Dependency{
			Name:  item,
			Level: 0,
		})
	}
	return &deps
}

// Returns a slice of the Dependency items currently held by this instance.
func (d *Dependencies) Slice() []*Dependency {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	s := []*Dependency{}
	for _, v := range d.dependencies {
		s = append(s, v)
	}
	return s
}

// Returns true if a Dependency of the same Name as dep is already present in the
// Dependencies collection.
func (d *Dependencies) Contains(dep Dependency) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.membership[dep.Name]
}

// Adds the given Dependency to the collection, returning an error if it already exists
// by name. Nil is returned if the Dependency is added successfully.
func (d *Dependencies) Add(dep *Dependency) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.membership[dep.Name] {
		return errors.New("The collection already contains a Dependency with this name")
	}
	d.membership[dep.Name] = true
	d.dependencies[dep.Name] = dep
	return nil
}

func (d *Dependencies) Len() int {
	return len(d.dependencies)
}

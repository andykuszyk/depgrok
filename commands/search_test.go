package commands

import (
	"testing"
	"sync"
	"strings"
	"path/filepath"

	"github.com/andykuszyk/depgrok/deps"
)

func TestSearchChildren_ShouldFindSimpleDependency(t *testing.T) {
	paralleliseSearches = false
	wg := sync.WaitGroup{}
	dependencies := deps.BuildDependencies(strings.Fields("dependency1"))
	searchChildren("", filepath.Join("..", "testdata"), dependencies, 0, &wg, []string{"*.md"}, make(chan repoCount, 100))
	wg.Wait()
	slice := dependencies.Slice()
	if len(slice) != 2 {
		t.Errorf("Expected there to be 2 dependencies, but there were %d", len(slice))
	}
	var dependency *deps.Dependency
	for _, dep := range slice {
		if dep.Name == "dependency1" {
			dependency = dep
			break
		}
	}
	if len(dependency.Repos) != 1 {
		t.Errorf("There should have been 1 repo for the dependency, but there were %d", len(dependency.Repos))
	}
	if !dependency.Repos["repo1"] {
		var key string
		for k, _ := range dependency.Repos {
			key = k
			break
		}
		t.Errorf("The repo should have been repo1, but it was %s", key)
	}
}


func TestStripExtension_WhenTwoExtensions(t *testing.T) {
	result := stripExtension("file.txt.txt")
	if result != "file.txt" {
		t.Errorf("Expected file.txt, but got %s", result)
	}
}

func TestStripExtension_WhenOneExtension(t *testing.T) {
	result := stripExtension("file.txt")
	if result != "file" {
		t.Errorf("Expected file, but got %s", result)
	}
}

func TestGetNextLoadingChar(t *testing.T) {
	secondsBetweenLoadingChars = 0.0
	char := getNextLoadingChar()
	if char != "/" {
		t.Errorf("Expected /, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "-" {
		t.Errorf("Expected -, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "\\" {
		t.Errorf("Expected \\, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "|" {
		t.Errorf("Expected |, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "/" {
		t.Errorf("Expected /, but got %s", char)
	}
}

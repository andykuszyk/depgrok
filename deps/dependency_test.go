package deps

import (
	"testing"
)

func TestDependencyAddRepo_ShouldAddRepo(t *testing.T) {
	sut := Dependency{}

	sut.AddRepo("repo")

	if len(sut.Repos) != 1 {
		t.Error("There should be one repo in the map")
	}
	if !sut.Repos["repo"] {
		t.Error("The value of repo in the map should be true")
	}
}

func TestDependencyMatches_ShouldMatchWhenTextContainsName(t *testing.T) {
	sut := Dependency{Name: "foo"}
	text := "foobar blah blah"

	result := sut.Matches(text)

	if !result {
		t.Error("Text should match name")
	}
}

func TestDependencyMatches_ShouldNotMatchWhenTextNotContainsName(t *testing.T) {
	sut := Dependency{Name: "foo"}
	text := "bar blah blah"

	result := sut.Matches(text)

	if result {
		t.Error("Text should not match name")
	}
}

func TestDependencyDependencyDiagram_ShouldReturnSimpleDiagram_WithNoParent(t *testing.T) {
	sut := Dependency{Name: "foo"}

	diagram := sut.DependencyDiagram("bar")

	if diagram.Text != "bar -> foo" {
		t.Error("Diagram text should be [repo] -> [dependency]")
	}
}

func TestDependencyDependencyDiagram_ShouldReturnDiagram_WithParent(t *testing.T) {
	parent := Dependency{Name: "go"}
	sut := Dependency{Name: "fer", Parent: &parent}

	diagram := sut.DependencyDiagram("bar")

	if diagram.Text != "bar -> fer -> go" {
		t.Errorf("Diagram text should be [repo] -> [parent] -> [dependency], but was %s", diagram.Text)
	}
}

func contains(slice []*Dependency, element string) bool {
	for _, dep := range slice {
		if dep.Name == element {
			return true
		}
	}
	return false
}

func TestDependenciesSlice_ShouldReturnSlice(t *testing.T) {
	sut := BuildDependencies([]string{"one", "two", "three"})

	slice := sut.Slice()

	if len(slice) != 3 {
		t.Errorf("The returned slice should have length 3, but instead it was %d", len(slice))
	}
	if !(contains(slice, "one") && contains(slice, "two") && contains(slice, "three")) {
		t.Errorf("Slice should contain [one, two, three], but it was: %v", slice)
	}
}

func TestDependenciesContains_ShouldReturnTrueWhenDependencyExists(t *testing.T) {
	sut := BuildDependencies([]string{"one", "two", "three"})

	result := sut.Contains(Dependency{Name: "one"})

	if !result {
		t.Error("Contains should return true when the dependency is contained within the collection")
	}
}

func TestDependenciesContains_ShouldReturnFalseWhenDependencyNotExists(t *testing.T) {
	sut := BuildDependencies([]string{"one", "two", "three"})

	result := sut.Contains(Dependency{Name: "four"})

	if result {
		t.Error("Contains should return false when the dependency is not contained within the collection")
	}
}

func TestDependenciesAdd_ShouldAddToCollection(t *testing.T) {
	sut := BuildDependencies([]string{"one"})
	err := sut.Add(&Dependency{Name: "two"})

	if err != nil {
		t.Error("No error should be returned when adding a new dependency")
	}
	if !sut.Contains(Dependency{Name: "two"}) {
		t.Error("Contains should return true for a Dependency that has been added using Add")
	}
	if !contains(sut.Slice(), "two") {
		t.Error("Slice should return a new Dependency after it has been added with Add")
	}
}

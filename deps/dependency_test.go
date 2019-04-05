package deps

import (
	"testing"
	"sort"

	"github.com/stretchr/testify/assert"
)

func TestDependencyDiagramKeys_Len(t *testing.T) {
	sut := dependencyDiagramKeys{
		items: []dependencyDiagramKey{
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo1",
				Level: 1,
			},
			dependencyDiagramKey {
				Name: "name2",
				Repo: "repo2",
				Level: 2,
			},
		},
	}

	actual := sut.Len()

	assert.Equal(t, 2, actual)
}

func TestDependencyDiagramKeys_CanSortByName(t *testing.T) {
	sut := dependencyDiagramKeys{
		items: []dependencyDiagramKey{
			dependencyDiagramKey {
				Name: "name2",
				Repo: "repo1",
				Level: 1,
			},
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo1",
				Level: 1,
			},
		},
	}

	sort.Sort(&sut)

	assert.Equal(t, "name1", sut.items[0].Name)
	assert.Equal(t, "name2", sut.items[1].Name)
}

func TestDependencyDiagramKeys_CanSortByRepo(t *testing.T) {
	sut := dependencyDiagramKeys{
		items: []dependencyDiagramKey{
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo2",
				Level: 1,
			},
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo1",
				Level: 1,
			},
		},
	}

	sort.Sort(&sut)

	assert.Equal(t, "repo1", sut.items[0].Repo)
	assert.Equal(t, "repo2", sut.items[1].Repo)
}

func TestDependencyDiagramKeys_CanSortByLevel(t *testing.T) {
	sut := dependencyDiagramKeys{
		items: []dependencyDiagramKey{
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo1",
				Level: 2,
			},
			dependencyDiagramKey {
				Name: "name1",
				Repo: "repo1",
				Level: 1,
			},
		},
	}

	sort.Sort(&sut)

	assert.Equal(t, 1, sut.items[0].Level)
	assert.Equal(t, 2, sut.items[1].Level)
}

func TestDependencyDiagramKeys_CanSortByRepoNameLevel(t *testing.T) {
	sut := dependencyDiagramKeys{
		items: []dependencyDiagramKey{
			dependencyDiagramKey {Name: "name2", Repo: "repo2", Level: 1},
			dependencyDiagramKey {Name: "name1", Repo: "repo1", Level: 1},
			dependencyDiagramKey {Name: "name1-ancestor", Repo: "repo1", Level: 2},
			dependencyDiagramKey {Name: "name1-ancestor", Repo: "repo2", Level: 2},
			dependencyDiagramKey {Name: "name2", Repo: "repo1", Level: 1},
			dependencyDiagramKey {Name: "name1", Repo: "repo2", Level: 1},
		},
	}

	sort.Sort(&sut)

	for _, item := range sut.items[:3] {
		assert.Equal(t, "repo1", item.Repo)
	}
	for _, item := range sut.items[3:] {
		assert.Equal(t, "repo2", item.Repo)
	}
	assert.Equal(t, "name1", sut.items[0].Name)
	assert.Equal(t, "name1-ancestor", sut.items[1].Name)
	assert.Equal(t, "name2", sut.items[2].Name)
	assert.Equal(t, "name1", sut.items[3].Name)
	assert.Equal(t, "name1-ancestor", sut.items[4].Name)
	assert.Equal(t, "name2", sut.items[5].Name)
}

func TestBuildDiagrams_WithTwoDepsTwoRepos(t *testing.T) {
	sut := BuildDependencies([]string{"dep1", "dep2"})
	sut.Slice()[0].AddRepo("repo1")
	sut.Slice()[1].AddRepo("repo2")
	sut.Slice()[1].Parent = sut.Slice()[0]
	sut.Slice()[1].Level = 1

	actual := sut.BuildDiagrams()

	if len(actual) != 2 {
		t.Errorf("Expected 2 diagrams, but found %d", len(actual))
	}
	t.Logf(actual[0].Text)
	if actual[0].Text != "repo1 -> dep1" {
		t.Errorf("Expected repo1 -> dep1, but got %s", actual[0].Text)
	}
	t.Logf(actual[1].Text)
	if actual[1].Text != "repo2 -> dep2 -> dep1" {
		t.Errorf("Expected repo2 -> dep2 -> dep1, but got %s", actual[1].Text)
	}
}

func TestBuildDiagrams_WithTwoDepsSharedRepo(t *testing.T) {
	sut := BuildDependencies([]string{"dep1", "dep2"})
	sut.Slice()[0].AddRepo("repo1")
	sut.Slice()[1].AddRepo("repo1")
	sut.Slice()[1].Parent = sut.Slice()[0]
	sut.Slice()[1].Level = 1

	actual := sut.BuildDiagrams()

	assert.Equal(t, 2, len(actual), "2 diagrams should be generated")
	if len(actual) < 2 {
		return
	}
	assert.Equal(t, "repo1 -> dep1", actual[0].Text)
	assert.Equal(t, "repo1 -> dep2 -> dep1", actual[1].Text)
}

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

func TestDependencyMatches_ShouldNotMatchWhenTextSurroundedByOtherChars(t *testing.T) {
	sut := Dependency{Name: "foo"}
	text := "spamfooeggs"

	result := sut.Matches(text)

	if result {
		t.Error("Text should not match when it is nested amongst other characters")
	}
}

func TestDependencyMatches_ShouldMatchWhenTextContainsName(t *testing.T) {
	sut := Dependency{Name: "foo"}
	text := "bla.foo bar blah blah"

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

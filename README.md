![](https://travis-ci.org/andykuszyk/depgrok.svg?branch=master)

# depgrok
A [very] simple tool for groking dependencies between multiple projects.

## Why depgrok?
This project was first written to try to understand how different database tables and stored procedures were being referenced around a larger code base.

## Installation
* `go get github.com/andykuszyk/depgrok`
* `cd $GO_PATH/github.com/andykuszyk/depgrok`
* `go install`

## Usage
### Cloning repos
To clone repos from an entire GitHub organisation into a new folder, use:

```
depgrok clone --org [github-org-name] --token [github-personal-access-token] --dir [destination-directory]
```

> `depgrok clone` uses SSH repo paths and assumes that the users SSH credentials have been uploaded to GitHub.

### Searching repos for dependencies
To search a directory of repos for links to a dependant reference, use:

```
depgrok search --deps [white-space-separated-dependancies] --dir [directory to search] --depth 1
```

> The depth flag is there to expand dependency chains beyond the default length of one, however this functionality is still in draft.

package main

import (
	"github.com/andykuszyk/depgrok/commands"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "depgrok"
	app.Version = "0.1"
	app.Usage = "Analyses a set of code repositories for references that depend on an input set."
	app.Commands = []cli.Command{
		{
			Name: "clone",
			Usage: "Clones git repositories for a given GitHub organisation to a given location. " +
				"Assumes the user has SSH access to the organisation provided in --org",
			Action: commands.Clone,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "org",
					Usage: "The GitHub organisation to clone repositories from",
				},
				cli.StringFlag{
					Name:  "dir",
					Usage: "The output directory in which to place the cloned repositories",
				},
				cli.StringFlag{
					Name:  "token",
					Usage: "The Personal Access Token required to authenticate with the GitHub API",
				},
			},
		},
		{
			Name: "search",
			Usage: "Searches a directory of code repositories for references to entities, returning " +
				"those repositories that match, directly or indirectly",
			Action: commands.Search,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "deps",
					Usage: "The dependencies to search for, provided as a white-space separated list",
				},
				cli.StringFlag{
					Name:  "dir",
					Usage: "The directory containing code repositories, in which to search",
				},
				cli.IntFlag{
					Name: "depth",
					Usage: "The depth of the dependency tree to construct. E.g. a value of 1 (the default)" +
						" will result in direct dependency relationships being found (X depends on Y, " +
						" or X -> Y). A value of 2 results in direct relationships, with one dependency" +
						" in common being found (X depends on Z, via Y, or X -> Y -> Z).",
					Value: 1,
				},
				cli.StringSliceFlag{
					Name:  "exclude",
					Usage: "A glob or file to exclude from the dependency search, e.g. *.md. If this option" +
						" is set, all files will be searched apart from those that match the glob. Cannot be" +
						" used in conjunction with --include",
				},
				cli.StringSliceFlag{
					Name:  "include",
					Usage: "A glob or file to include from the dependency search, e.g. *.cs. If this option" +
						" is set, only files that match the glob will be searched, at the exclusion of all " +
						"others. Cannot be used in conjunction with --exclude",
				},
				cli.BoolFlag{
					Name: "debug",
					Usage: "Prints additional debug information to stderr",
				},
			},
		},
	}

	app.Run(os.Args)
}

package main

import (
    "os"
    "github.com/urfave/cli"
    "github.com/carfinance247/depgrok/commands"
)

func main() {
    app := cli.NewApp()
    app.Name = "depgrok"
    app.Version = "0.1"
    app.Usage = "Analyses a set of code repositories for references that depend on an input set."
    app.Commands = []cli.Command{
        {
            Name: "clone",
            Usage: "Clones git repositories for a given GitHub organisation to a given location. "+
                    "Assumes the user has SSH access to the organisation provided in --org",
            Action: commands.Clone,
            Flags: []cli.Flag {
                cli.StringFlag {
                    Name: "org",
                    Usage: "The GitHub organisation to clone repositories from",
                },
                cli.StringFlag {
                    Name: "dir",
                    Usage: "The output directory in which to place the cloned repositories",
                },
                cli.StringFlag {
                    Name: "token",
                    Usage: "The Personal Access Token required to authenticate with the GitHub API",
                },
            },
        },
        {
            Name: "search",
            Usage: "Searches a directory of code repositories for references to entities, returning "+
                   "those repositories that match, directly or indirectly",
            Action: commands.Search,
            Flags: []cli.Flag {
                cli.StringFlag {
                    Name: "deps",
                    Usage: "The dependencies to search for, provided as a white-space separated list",
                },
                cli.StringFlag {
                    Name: "dir",
                    Usage: "The directory containing code repositories, in which to search",
                },
            },
        },
    }

    app.Run(os.Args)
}

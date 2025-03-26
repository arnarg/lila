package main

import (
	"log"
	gos "os"

	"github.com/arnarg/lila/cmd/lila/build"
	"github.com/arnarg/lila/cmd/lila/os"
	"github.com/arnarg/lila/cmd/lila/shell"
	"github.com/urfave/cli/v2"
)

var version = "unknown"

func main() {
	app := &cli.App{
		Name:        "lila",
		Version:     version,
		Description: "Alternative CLI for nilla projects.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Set log level to verbose",
			},
		},
		Commands: cli.Commands{
			build.Command,
			os.Command,
			shell.Command,
		},
	}

	if err := app.Run(gos.Args); err != nil {
		log.Fatal(err)
	}
}

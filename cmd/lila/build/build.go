package build

import (
	"context"
	"fmt"
	"log"

	"github.com/arnarg/lila/internal/nix"
	"github.com/arnarg/lila/internal/tui"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:        "build",
	Aliases:     []string{"b"},
	Usage:       "Build a package",
	Description: "Builds a package defined in a nilla project",
	Args:        true,
	ArgsUsage:   "[package name]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-link",
			Usage: "Do not create symlinks to the build results",
		},
		&cli.BoolFlag{
			Name:  "print-out-paths",
			Usage: "Print the resulting output paths",
		},
		&cli.StringFlag{
			Name:    "out-link",
			Aliases: []string{"o"},
			Usage:   "Use path as prefix for the symlinks to the build results",
		},
	},
	Action: run,
}

func run(ctx *cli.Context) error {
	args := ctx.Args()
	name := args.First()

	if name == "" {
		name = "default"
	}

	// Get current system
	system, err := nix.CurrentSystem()
	if err != nil {
		return err
	}

	// Build args for nix build
	nargs := []string{
		"-f", "nilla.nix",
		fmt.Sprintf("packages.%s.result.%s", name, system),
	}

	if ctx.Bool("no-link") {
		nargs = append(nargs, "--no-link")
	}
	if ctx.String("out-link") != "" {
		nargs = append(nargs, "--out-link", ctx.String("out-link"))
	}

	// Create a build TUI reporter
	reporter := tui.NewBuildReporter(ctx.Bool("verbose"))

	// Run nix build
	out, err := nix.Command("build").
		Args(nargs).
		Reporter(reporter).
		Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Print out path, if wanted
	if ctx.Bool("print-out-paths") {
		fmt.Println(string(out))
	}

	return nil
}

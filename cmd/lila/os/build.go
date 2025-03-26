package os

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var buildCmd = &cli.Command{
	Name:        "build",
	Usage:       "Build NixOS configuration",
	Description: "Build NixOS configuration",
	Args:        true,
	ArgsUsage:   "[system name]",
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
	Action: runBuild,
}

func runBuild(ctx *cli.Context) error {
	args := ctx.Args()
	name, err := inferName(args.First())
	if err != nil {
		return err
	}

	// Build extra args
	eargs := []string{}
	if ctx.Bool("--no-link") {
		eargs = append(eargs, "--no-link")
	}
	if ctx.String("out-link") != "" {
		eargs = append(eargs, "--out-link", ctx.String("out-link"))
	}

	// Run nix build
	out, err := buildToplevel(name, eargs, ctx.Bool("verbose"))
	if err != nil {
		return err
	}

	// Print out path, if wanted
	if ctx.Bool("print-out-paths") {
		fmt.Println(string(out))
	}

	return nil
}

package shell

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/arnarg/lila/internal/nix"
	"github.com/arnarg/lila/internal/tui"
	"github.com/urfave/cli/v2"
)

const NIX_SOURCED_VAR = "__ETC_PROFILE_NIX_SOURCED"

var Command = &cli.Command{
	Name:        "shell",
	Aliases:     []string{"s"},
	Usage:       "Run a nix shell",
	Description: "Builds and runs a shell defined in a nilla project",
	Args:        true,
	ArgsUsage:   "[shell name]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "command",
			Aliases: []string{"c"},
			Usage:   "Command and arguments to be executed, defaulting to $SHELL",
			EnvVars: []string{"SHELL"},
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
	attr := fmt.Sprintf("shells.%s.result.%s", name, system)
	nargs := []string{
		"-f", "nilla.nix", attr, "--no-link",
	}

	// Create a build TUI reporter
	reporter := tui.NewBuildReporter(ctx.Bool("verbose"))

	// Run nix build
	_, err = nix.Command("build").
		Args(nargs).
		Reporter(reporter).
		Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create nix-shell command
	sargs := []string{
		"nilla.nix",
		"--attr", attr,
		"--quiet",
	}

	if ctx.String("command") != "" {
		sargs = append(sargs, "--command", ctx.String("command"))
	}

	shell := exec.Command("nix-shell", sargs...)

	// Inherit environment
	shell.Env = append(os.Environ(), fmt.Sprintf("%s=1", NIX_SOURCED_VAR))

	// Plug pipes
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr

	// Run nix-shell
	if err := shell.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			os.Exit(eerr.ExitCode())
		} else {
			return err
		}
	}

	return nil
}

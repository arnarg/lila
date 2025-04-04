package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

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

	// Run nix build
	_, err = nix.Command("build").
		Args(nargs).
		Reporter(tui.NewBuildReporter(ctx.Bool("verbose"))).
		Run(context.Background())
	if err != nil {
		return err
	}

	// Create nix-shell arg list
	sargs := []string{"nix-shell", "nilla.nix", "--attr", attr, "--quiet"}

	if ctx.String("command") != "" {
		sargs = append(sargs, "--command", ctx.String("command"))
	}

	// Copy current environment with NIX_SOURCED_VAR set
	senv := append(os.Environ(), fmt.Sprintf("%s=1", NIX_SOURCED_VAR))

	// Find full path to nix-shell
	spath, err := exec.LookPath("nix-shell")
	if err != nil {
		return err
	}

	return syscall.Exec(spath, sargs, senv)
}

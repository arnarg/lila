package os

import (
	"github.com/urfave/cli/v2"
)

var testCmd = &cli.Command{
	Name:        "test",
	Usage:       "Build NixOS configuration and activate it",
	Description: "Build NixOS configuration and activate it",
	Args:        true,
	ArgsUsage:   "[system name]",
	Action:      runTest,
}

func runTest(ctx *cli.Context) error {
	args := ctx.Args()
	name, err := inferName(args.First())
	if err != nil {
		return err
	}

	// Run nix build
	out, err := buildToplevel(name, []string{"--no-link"}, ctx.Bool("verbose"))
	if err != nil {
		return err
	}

	// Activate toplevel
	return activateToplevel(string(out))
}

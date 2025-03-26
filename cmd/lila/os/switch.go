package os

import (
	"github.com/urfave/cli/v2"
)

var switchCmd = &cli.Command{
	Name:        "switch",
	Usage:       "Build NixOS configuration, activate it and make it the boot default",
	Description: "Build NixOS configuration, activate it and make it the boot default",
	Args:        true,
	ArgsUsage:   "[system name]",
	Action:      runSwitch,
}

func runSwitch(ctx *cli.Context) error {
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
	if err := activateToplevel(string(out)); err != nil {
		return err
	}

	// Set boot for toplevel
	return setBootToplevel(string(out))
}

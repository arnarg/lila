package os

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/arnarg/lila/internal/nix"
	"github.com/arnarg/lila/internal/tui"
	"github.com/urfave/cli/v2"
)

type subCmd int

const (
	subCmdBuild subCmd = iota
	subCmdTest
	subCmdBoot
	subCmdSwitch
)

const SYSTEM_PROFILE = "/nix/var/nix/profiles/system"
const CURRENT_PROFILE = "/run/current-system"

var Command = &cli.Command{
	Name:        "os",
	Usage:       "NixOS operations",
	Description: "Reimplementation of nixos-rebuild",
	Subcommands: []*cli.Command{
		// Build
		{
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
			Action: func(ctx *cli.Context) error {
				return run(ctx, subCmdBuild)
			},
		},

		// Test
		{
			Name:        "test",
			Usage:       "Build NixOS configuration and activate it",
			Description: "Build NixOS configuration and activate it",
			Args:        true,
			ArgsUsage:   "[system name]",
			Action: func(ctx *cli.Context) error {
				return run(ctx, subCmdTest)
			},
		},

		// Boot
		{
			Name:        "boot",
			Usage:       "Build NixOS configuration and make it the boot default",
			Description: "Build NixOS configuration and make it the boot default",
			Args:        true,
			ArgsUsage:   "[system name]",
			Action: func(ctx *cli.Context) error {
				return run(ctx, subCmdBoot)
			},
		},

		// Switch
		{
			Name:        "switch",
			Usage:       "Build NixOS configuration, activate it and make it the boot default",
			Description: "Build NixOS configuration, activate it and make it the boot default",
			Args:        true,
			ArgsUsage:   "[system name]",
			Action: func(ctx *cli.Context) error {
				return run(ctx, subCmdSwitch)
			},
		},
	},
}

func printSection(text string) {
	fmt.Fprintf(os.Stderr, "\033[32m>\033[0m %s\n", text)
}

func inferName(name string) (string, error) {
	if name == "" {
		hn, err := os.Hostname()
		if err != nil {
			return "", err
		}
		return hn, nil
	}
	return name, nil
}

func run(ctx *cli.Context, sc subCmd) error {
	// Try to infer name of the NixOS system
	name, err := inferName(ctx.Args().First())
	if err != nil {
		return err
	}

	// Attribute of NixOS configuration's toplevel
	attr := fmt.Sprintf("systems.nixos.%s.result.config.system.build.toplevel", name)

	//
	// NixOS configuration build
	//
	// Build args for nix build
	nargs := []string{"-f", "nilla.nix", attr}

	// Add extra args depending on the sub command
	if sc == subCmdBuild {
		if ctx.Bool("no-link") {
			nargs = append(nargs, "--no-link")
		}
		if ctx.String("out-link") != "" {
			nargs = append(nargs, "--out-link", ctx.String("out-link"))
		}
	} else {
		// All sub-commands except build should not
		// create a result link
		nargs = append(nargs, "--no-link")
	}

	// Run nix build
	printSection("Building configuration")
	out, err := nix.Command("build").
		Args(nargs).
		Reporter(tui.NewBuildReporter(ctx.Bool("verbose"))).
		Run(context.Background())
	if err != nil {
		return err
	}

	//
	// Run generation diff using nvd
	//
	fmt.Fprintln(os.Stderr)
	printSection("Comparing changes")

	// Run nvd diff
	diff := exec.Command("nvd", "diff", CURRENT_PROFILE, string(out))
	diff.Stderr = os.Stderr
	diff.Stdout = os.Stderr
	if err := diff.Run(); err != nil {
		return err
	}

	//
	// Activate NixOS configuration
	//
	if sc == subCmdTest || sc == subCmdSwitch {
		fmt.Fprintln(os.Stderr)
		printSection("Activating configuration")

		// Run switch_to_configuration
		switchp := fmt.Sprintf("%s/bin/switch-to-configuration", out)
		switchc := exec.Command("sudo", switchp, "test")
		switchc.Stderr = os.Stderr
		switchc.Stdout = os.Stdout

		if err := switchc.Run(); err != nil {
			// TODO: maybe we should continue during switch
			return err
		}
	}

	//
	// Set NixOS configuration in bootloader
	//
	if sc == subCmdBoot || sc == subCmdSwitch {
		// Set profile
		_, err := nix.Command("build").
			Args([]string{
				"--no-link",
				"--profile", SYSTEM_PROFILE,
				string(out),
			}).
			Privileged(true).
			Run(context.Background())
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stderr)
		printSection("Adding configuration to bootloader")

		// Run switch_to_configuration
		switchp := fmt.Sprintf("%s/bin/switch-to-configuration", out)
		switchc := exec.Command("sudo", switchp, "boot")
		switchc.Stderr = os.Stderr
		switchc.Stdout = os.Stdout

		return switchc.Run()
	}

	return nil
}

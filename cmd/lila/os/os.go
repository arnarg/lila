package os

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

const SYSTEM_PROFILE = "/nix/var/nix/profiles/system"
const CURRENT_PROFILE = "/run/current-system"

var Command = &cli.Command{
	Name:        "os",
	Usage:       "NixOS operations",
	Description: "Reimplementation of nixos-rebuild",
	Subcommands: []*cli.Command{
		buildCmd,
		bootCmd,
		testCmd,
		switchCmd,
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

func buildToplevel(name string, extraArgs []string, verbose bool) ([]byte, error) {
	printSection("Building NixOS configuration")

	// Build args for nix build
	nargs := []string{
		"-f", "nilla.nix",
		fmt.Sprintf("systems.nixos.%s.result.config.system.build.toplevel", name),
	}
	nargs = append(nargs, extraArgs...)

	// Create a build TUI reporter
	reporter := tui.NewBuildReporter(verbose)

	// Run nix build
	out, err := nix.Command("build").
		Args(nargs).
		Reporter(reporter).
		Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr)
	printSection("Comparing changes")

	// Run nvd diff
	diff := exec.Command("nvd", "diff", CURRENT_PROFILE, string(out))
	diff.Stderr = os.Stderr
	diff.Stdout = os.Stderr
	if err := diff.Run(); err != nil {
		return nil, err
	}

	return out, nil
}

func activateToplevel(out string) error {
	fmt.Fprintln(os.Stderr)
	printSection("Activating configuration")

	// Run switch_to_configuration
	switchp := fmt.Sprintf("%s/bin/switch-to-configuration", out)
	switchc := exec.Command("sudo", switchp, "test")
	switchc.Stderr = os.Stderr
	switchc.Stdout = os.Stdout

	return switchc.Run()
}

func setBootToplevel(out string) error {
	// Set profile
	_, err := nix.Command("build").
		Args([]string{
			"--no-link",
			"--profile", SYSTEM_PROFILE,
			out,
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

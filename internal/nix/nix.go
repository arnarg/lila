package nix

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// CurrentSystem returns `builtins.currentSystem` from `nix eval`.
func CurrentSystem() (string, error) {
	sys, err := exec.Command(
		"nix", "eval",
		"--expr", "builtins.currentSystem", "--raw", "--impure",
	).Output()
	if err != nil {
		return "", err
	}

	return string(sys), nil
}

type NixCommand struct {
	cmd  string
	args []string

	privileged bool

	reporter ProgressReporter
}

func Command(cmd string) NixCommand {
	return NixCommand{
		cmd: cmd,
	}
}

func (c NixCommand) Args(args []string) NixCommand {
	c.args = args
	return c
}

func (c NixCommand) Privileged(privileged bool) NixCommand {
	c.privileged = privileged
	return c
}

func (c NixCommand) Reporter(reporter ProgressReporter) NixCommand {
	c.reporter = reporter
	return c
}

func (c NixCommand) Run(ctx context.Context) ([]byte, error) {
	cmd := "nix"
	args := []string{}

	// Check if we need to run with sudo
	if c.privileged {
		cmd = "sudo"
		args = append(args, "nix")
	}

	// Append rest of arguments
	args = append(args, c.cmd, "--print-out-paths")
	args = append(args, c.args...)

	if c.reporter != nil {
		return c.runWithReporter(ctx, cmd, args)
	}

	return c.runStdout(ctx, cmd, args)
}

func (c NixCommand) runStdout(ctx context.Context, cmd string, args []string) ([]byte, error) {
	// Create nix command
	nixc := exec.CommandContext(ctx, cmd, args...)

	// Create a buffer to capture nix's stdout
	b := &bytes.Buffer{}

	// Plug stdout and stderr
	nixc.Stdout = b
	nixc.Stderr = os.Stderr

	// Run nix command
	if err := nixc.Run(); err != nil {
		return nil, err
	}

	return bytes.TrimSpace(b.Bytes()), nil
}

func (c NixCommand) runWithReporter(ctx context.Context, cmd string, args []string) (res []byte, err error) {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sctx, stop := signal.NotifyContext(cctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Add internal-json format flags
	args = append(args, "--log-format", "internal-json", "-v")

	// Create nix command
	nixc := exec.CommandContext(sctx, cmd, args...)

	// Create a buffer to capture nix's stdout
	b := &bytes.Buffer{}
	nixc.Stdout = b

	// Get stderr pipe from nix command
	stderr, err := nixc.StderrPipe()
	if err != nil {
		return
	}

	// Start nix command
	if err = nixc.Start(); err != nil {
		return
	}

	// Run progress reporter
	var perr error
	if perr = c.reporter.Run(sctx, NewProgressDecoder(stderr)); perr != nil {
		cancel()
	}

	// Wait for nix command and decoder
	var cerr error
	if cerr = nixc.Wait(); cerr != nil {
		if cerr == context.Canceled {
			cerr = nil
		} else {
			if cerr.Error() == "signal: killed" {
				cerr = nil
			}
		}
	}

	// Set error
	if perr != nil {
		err = perr
		return
	} else if cerr != nil {
		err = cerr
		return
	}

	res = bytes.TrimSpace(b.Bytes())
	return
}

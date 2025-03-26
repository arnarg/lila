package main

import (
	"context"
	"log"
	"os"

	"github.com/arnarg/lila/internal/nix"
	"github.com/arnarg/lila/internal/tui"
)

func main() {
	// Run progress reporter
	err := tui.NewBuildReporter(false).Run(
		context.Background(),
		nix.NewProgressDecoder(os.Stdin),
	)
	if err != nil {
		log.Fatal(err)
	}
}

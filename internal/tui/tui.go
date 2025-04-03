package tui

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/arnarg/lila/internal/nix"
	"github.com/arnarg/lila/internal/util"
	tea "github.com/charmbracelet/bubbletea"
)

type tuiModel interface {
	tea.Model
	error() error
}

func runTUIModel(ctx context.Context, init tuiModel, decoder *nix.ProgressDecoder) error {
	var wg sync.WaitGroup

	p := tea.NewProgram(
		init,
		// Signal handling is done outside and
		// cancels ctx
		tea.WithoutSignalHandler(),
		// Output on stderr so that --print-out-paths
		// can print on stdout
		tea.WithOutput(os.Stderr),
		// Signal handling works outside of bubbletea
		// when input is nil
		tea.WithInput(nil),
		// Seems enough
		tea.WithFPS(30),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for ev := range decoder.Events {
			// Check if context has been cancelled
			select {
			case <-ctx.Done():
				p.Quit()
				return
			default:
			}

			// Send event to program
			p.Send(ev)
		}

		p.Quit()
	}()

	// Run bubbletea program
	m, err := p.Run()
	if err != nil {
		return err
	}

	// Wait for waitgroup
	wg.Wait()

	return m.(tuiModel).error()
}

type progress struct {
	done     int
	expected int
	running  int
}

type progresses map[int64]progress

func (p progresses) count() int {
	return len(p)
}

func (p progresses) totalDone() int {
	total := 0
	for _, prog := range p {
		total += prog.done
	}
	return total
}

func (p progresses) totalExpected() int {
	total := 0
	for _, prog := range p {
		total += prog.expected
	}
	return total
}

func (p progresses) totalRunning() int {
	total := 0
	for _, prog := range p {
		total += prog.running
	}
	return total
}

type copy struct {
	name  string
	done  int64
	total int64
}

func (c *copy) String() string {
	if c.total > 0 {
		total, unit := util.ConvertBytes(c.total)
		done := util.ConvertBytesToUnit(c.done, unit)

		return fmt.Sprintf("%s [%.2f/%.2f %s]", c.name, done, total, unit)
	}
	return c.name
}

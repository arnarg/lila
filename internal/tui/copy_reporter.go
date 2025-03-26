package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/arnarg/lila/internal/nix"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CopyReporter struct {
	verbose bool
}

func NewCopyReporter(verbose bool) *CopyReporter {
	return &CopyReporter{verbose}
}

func (r *CopyReporter) Run(ctx context.Context, decoder *nix.ProgressDecoder) error {
	return runTUIModel(ctx, initCopyModel(r.verbose), decoder)
}

type copyModel struct {
	spinner spinner.Model

	w, h int

	verbose     bool
	initialized bool

	copyPathsProgs progresses

	copies    map[int64]*copy
	transfers map[int64]int64

	lastMsg string

	err error
}

func initCopyModel(verbose bool) copyModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return copyModel{
		verbose:        verbose,
		spinner:        s,
		copyPathsProgs: map[int64]progress{},
		copies:         map[int64]*copy{},
		transfers:      map[int64]int64{},
		lastMsg:        "Initializing...",
	}
}

func (m copyModel) error() error {
	return m.err
}

func (m copyModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m copyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case nix.Event:
		return m.handleEvent(msg)
	}

	return m, nil
}

func (m copyModel) handleEvent(ev nix.Event) (tea.Model, tea.Cmd) {
	switch ev.Action() {
	case nix.ActionTypeStart:
		return m.handleStartEvent(ev)
	case nix.ActionTypeStop:
		event := ev.(nix.StopEvent)
		return m.handleStopEvent(event)
	case nix.ActionTypeResult:
		return m.handleResultEvent(ev)
	case nix.ActionTypeMessage:
		event := ev.(nix.MessageEvent)

		// error
		if event.Level == 0 {
			m.err = errors.New(event.Text)
			return m, tea.Quit
		}

		// Just display the message
		if m.verbose {
			return m, tea.Printf("%s", event.Text)
		} else {
			m.lastMsg = event.Text
		}
	}

	return m, nil
}

func (m copyModel) handleStartEvent(ev nix.Event) (tea.Model, tea.Cmd) {
	switch ev := ev.(type) {
	case nix.StartCopyPathsEvent:
		m.copyPathsProgs[ev.ID] = progress{}
		if !m.initialized {
			m.initialized = true
		}
		return m, nil

	case nix.StartCopyPathEvent:
		m.copies[ev.ID] = &copy{name: ev.Path}

		if m.verbose {
			return m, tea.Println(ev.Text)
		}
		return m, nil

	case nix.StartFileTransferEvent:
		if _, ok := m.copies[ev.Parent]; !ok {
			return m, nil
		}

		m.transfers[ev.ID] = ev.Parent
		return m, nil
	}

	return m, nil
}

func (m copyModel) handleStopEvent(ev nix.StopEvent) (tea.Model, tea.Cmd) {
	// Check if it's a copy
	if c, ok := m.copies[ev.ID]; ok {
		// Remove from copies map
		delete(m.copies, ev.ID)
		if m.verbose {
			m.lastMsg = ""
			return m, tea.Printf(
				"%s %s",
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")).
					SetString("✓").
					String(),
				c.name,
			)
		}
	}

	// Finally we want to also clean up transfer parent mapping
	if _, ok := m.transfers[ev.ID]; ok {
		// Remove parent mapping
		delete(m.transfers, ev.ID)
	}

	// Clear last message if all builds and downloads have stopped,
	// but only after initialization
	if m.initialized && len(m.copies) < 1 {
		m.lastMsg = ""
	}

	return m, nil
}

func (m copyModel) handleResultEvent(ev nix.Event) (tea.Model, tea.Cmd) {
	switch ev := ev.(type) {
	case nix.ResultProgressEvent:
		// Check if the event ID is a CopyPaths event
		if p, ok := m.copyPathsProgs[ev.ID]; ok {
			p.done = int(ev.Done)
			p.expected = int(ev.Expected)
			p.running = ev.Running
			m.copyPathsProgs[ev.ID] = p
			return m, nil
		}

		// Otherwise we check if it's a transfer
		var c *copy
		if co, ok := m.copies[ev.ID]; ok {
			c = co
		} else if parent, ok := m.transfers[ev.ID]; ok {
			co, ok := m.copies[parent]
			if !ok {
				return m, nil
			}
			c = co
		}

		if c != nil {
			c.done = ev.Done
			c.total = ev.Expected

			m.lastMsg = c.String()
		}
	}
	return m, nil
}

func (m copyModel) View() string {
	if m.err != nil {
		return ""
	}

	if !m.initialized {
		return m.uninitializedView()
	}

	return m.progressView()
}

func (m copyModel) uninitializedView() string {
	return fmt.Sprintf("%s%s\n", m.spinner.View(), m.lastMsg)
}

func (m copyModel) progressView() string {
	strb := &strings.Builder{}

	if m.lastMsg != "" {
		width := m.w - lipgloss.Width(m.spinner.View())
		msg := m.lastMsg
		if len(msg) > width {
			p := "..."
			l := (len(msg) - width) + len(p)
			msg = fmt.Sprintf("%s%s", p, msg[l:])
		}
		strb.WriteString(
			fmt.Sprintf("%s%s\n", m.spinner.View(), msg),
		)
	} else {
		strb.WriteString("\n")
	}

	transfers := fmtTransfers(m)

	hdr := lipgloss.NewStyle().
		Bold(true).
		Width(lipgloss.Width(transfers)).
		SetString("Transfers:").
		String()

	strb.WriteString(fmt.Sprintf("%s\n", hdr))
	strb.WriteString(fmt.Sprintf("%s\n", transfers))

	return strb.String()
}

func fmtTransfers(m copyModel) string {
	running := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		SetString(fmt.Sprintf("↑ %d", m.copyPathsProgs.totalRunning())).
		String()

	done := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		SetString(fmt.Sprintf("✓ %d", m.copyPathsProgs.totalDone())).
		String()

	remaining := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		SetString(
			fmt.Sprintf(
				"⧗ %d", m.copyPathsProgs.totalExpected()-m.copyPathsProgs.totalDone(),
			),
		).
		String()

	return fmt.Sprintf("%s | %s | %s", running, done, remaining)
}

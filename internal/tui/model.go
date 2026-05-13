package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/firfircelik/libkill/internal/db"
	"github.com/firfircelik/libkill/internal/scanner"
)

type state int

const (
	stateScanning state = iota
	stateResults
	stateCleaning
	stateDone
)

type Model struct {
	store    *db.Store
	state    state
	spinner  spinner.Model
	viewport viewport.Model
	styles   Styles

	scanners   []scanner.Scanner
	results    []scanner.Result
	selected   map[int]bool
	cursor     int
	statusMsg  string
	err        error
	cleaned    int
	width      int
	height     int
}

type scanDoneMsg struct {
	results []scanner.Result
	err     error
}

type cleanDoneMsg struct {
	count int
	err   error
}

func New(store *db.Store) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)

	return &Model{
		store:    store,
		state:    stateScanning,
		spinner:  s,
		viewport: vp,
		styles:   DefaultStyles(),
		scanners: []scanner.Scanner{
			scanner.NewNPMScanner(store),
			scanner.NewPIPScanner(store),
		},
		selected: make(map[int]bool),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startScan(),
	)
}

func (m *Model) startScan() tea.Cmd {
	return func() tea.Msg {
		var all []scanner.Result
		ctx := context.Background()

		for _, s := range m.scanners {
			matches, _, err := s.Scan(ctx)
			if err != nil {
				return scanDoneMsg{err: fmt.Errorf("%s: %w", s.Name(), err)}
			}
			all = append(all, matches...)
		}

		var matches []scanner.Result
		for _, r := range all {
			if r.Threat.Package != "" {
				matches = append(matches, r)
			}
		}

		return scanDoneMsg{results: matches}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6

	case spinner.TickMsg:
		if m.state == stateScanning {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case scanDoneMsg:
		m.state = stateResults
		m.results = msg.results
		if msg.err != nil {
			m.err = msg.err
		}
		if len(m.results) > 0 {
			m.selected[0] = true
		}
		return m, nil

	case cleanDoneMsg:
		m.state = stateDone
		m.cleaned = msg.count
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
	}

	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == stateScanning {
			return m, tea.Quit
		}
		if m.state == stateDone || len(m.results) == 0 {
			return m, tea.Quit
		}
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}

	case " ":
		if m.state == stateResults && len(m.results) > 0 {
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
		}

	case "enter":
		if m.state == stateResults && len(m.selected) > 0 {
			m.state = stateCleaning
			return m, m.cleanSelected()
		}

	case "a":
		if m.state == stateResults && len(m.results) > 0 {
			for i := range m.results {
				m.selected[i] = true
			}
		}

	case "n":
		if m.state == stateResults && len(m.results) > 0 {
			m.selected = make(map[int]bool)
		}
	}

	return m, nil
}

func (m *Model) cleanSelected() tea.Cmd {
	selected := m.selected
	results := m.results

	return func() tea.Msg {
		count := 0
		for i := range selected {
			r := results[i]
			var err error
			switch r.Ecosystem {
			case "npm":
				var c *exec.Cmd
				if strings.Contains(r.Location, "global") {
					c = exec.CommandContext(context.Background(), "npm", "uninstall", "-g", r.Package)
				} else {
					c = exec.CommandContext(context.Background(), "npm", "uninstall", r.Package)
				}
				c.Stdout = os.Stderr
				c.Stderr = os.Stderr
				err = c.Run()
			case "pip":
				c := exec.CommandContext(context.Background(), "pip3", "uninstall", "-y", r.Package)
				c.Stdout = os.Stderr
				c.Stderr = os.Stderr
				err = c.Run()
			}
			if err != nil {
				return cleanDoneMsg{count: count, err: fmt.Errorf("%s: %w", r.Package, err)}
			}
			count++
		}
		return cleanDoneMsg{count: count}
	}
}

func (m *Model) View() string {
	switch m.state {
	case stateScanning:
		return m.viewScanning()
	case stateResults:
		return m.viewResults()
	case stateCleaning:
		return m.viewCleaning()
	case stateDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

func (m *Model) viewScanning() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.styles.Title.Render("LibKill")+"\n\n"+
			m.spinner.View()+" Scanning for compromised packages...",
	)
}

func (m *Model) viewResults() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)) + "\nPress q to quit"
	}

	if len(m.results) == 0 {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.styles.Success.Render("No compromised packages found!"),
		)
	}

	var sb strings.Builder
	sb.WriteString(m.styles.Title.Render("Compromised Packages") + "\n\n")

	for i, r := range m.results {
		cursor := "  "
		checked := " "
		if i == m.cursor {
			cursor = "> "
		}
		if m.selected[i] {
			checked = "x"
		}

		line := fmt.Sprintf("%s[%s] %s@%s (%s)",
			cursor, checked, r.Package, r.Version, r.Location)
		style := m.styles.Item
		if i == m.cursor {
			style = m.styles.SelectedItem
		}
		sb.WriteString(style.Render(line) + "\n")
	}

	sb.WriteString("\n" + m.styles.Help.Render(
		"↑/↓: navigate  space: select  a: all  n: none  enter: clean  q: quit",
	))

	if len(m.selected) > 0 {
		sb.WriteString("\n" + m.styles.Warning.Render(
			fmt.Sprintf("%d package(s) selected for removal", len(m.selected)),
		))
	}

	return sb.String()
}

func (m *Model) viewCleaning() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.spinner.View()+" Removing compromised packages...",
	)
}

func (m *Model) viewDone() string {
	msg := fmt.Sprintf("Cleaned %d compromised package(s)!", m.cleaned)
	if m.err != nil {
		msg += fmt.Sprintf("\nWarning: %v", m.err)
	}
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.styles.Success.Render(msg)+"\n\nPress q to quit",
	)
}

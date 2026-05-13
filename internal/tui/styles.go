package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title        lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1),

		Item: lipgloss.NewStyle().
			PaddingLeft(2),

		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("205")).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
	}
}

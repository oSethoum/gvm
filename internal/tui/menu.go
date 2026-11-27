package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	cmd  string
	desc string
}

func (m menuItem) Title() string       { return m.cmd }
func (m menuItem) Description() string { return m.desc }
func (m menuItem) FilterValue() string { return m.cmd }

type menuModel struct {
	list     list.Model
	selected string
	quit     bool
}

func newMenuModel() menuModel {
	items := []list.Item{
		menuItem{"install", "Install a Go version"},
		menuItem{"use", "Switch active Go version"},
		menuItem{"list", "List installed/available versions"},
		menuItem{"uninstall", "Uninstall a Go version"},
		menuItem{"current", "Show active version"},
		menuItem{"available", "List available versions (remote)"},
		menuItem{"env", "Print environment variables"},
		menuItem{"move", "Move gvm to a new directory"},
		menuItem{"init", "Shell integration snippet"},
		menuItem{"doctor", "Diagnose setup"},
		menuItem{"quit", "Exit"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 40, 14)
	l.Title = "gvm"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	return menuModel{list: l}
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			m.quit = true
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			if it, ok := m.list.SelectedItem().(menuItem); ok {
				m.selected = it.cmd
				if m.selected == "quit" {
					m.quit = true
				}
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	return m.list.View()
}

// RunMenu shows an interactive menu and returns the selected command.
func RunMenu() (string, error) {
	p := tea.NewProgram(newMenuModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	mm := m.(menuModel)
	if mm.quit {
		return "", nil
	}
	return mm.selected, nil
}

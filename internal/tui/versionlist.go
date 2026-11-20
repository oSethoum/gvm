package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type versionItem struct {
	version string
	status  string
}

func (v versionItem) Title() string       { return v.version }
func (v versionItem) Description() string { return v.status }
func (v versionItem) FilterValue() string  { return v.version }

type versionListModel struct {
	input    textinput.Model
	list     list.Model
	all      []list.Item
	selected string
	quit     bool
}

func newVersionListModel(items []list.Item) versionListModel {
	ti := textinput.New()
	ti.Placeholder = "Search versions..."
	ti.CharLimit = 64
	ti.Width = 40
	ti.Focus()

	l := list.New(items, list.NewDefaultDelegate(), 60, 16)
	l.Title = "Go versions"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return versionListModel{input: ti, list: l, all: items}
}

func (m versionListModel) Init() tea.Cmd { return textinput.Blink }

func (m versionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// if input focused and has text, clear; else quit
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.filter("")
				return m, nil
			}
			m.quit = true
			return m, tea.Quit
		case "enter":
			if it, ok := m.list.SelectedItem().(versionItem); ok {
				m.selected = it.version
				return m, tea.Quit
			}
		case "up", "k", "down", "j":
			m.input.Blur()
			var c tea.Cmd
			m.list, c = m.list.Update(msg)
			return m, c
		case "tab":
			// toggle focus between input and list
			if m.input.Focused() {
				m.input.Blur()
			} else {
				m.input.Focus()
			}
			return m, nil
		}
	}
	// if not focused and typing, focus input
	if !m.input.Focused() {
		if km, ok := msg.(tea.KeyMsg); ok && len(km.Runes) > 0 {
			m.input.Focus()
		}
	}
	// update input
	var cmd tea.Cmd
	if m.input.Focused() {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		m.filter(m.input.Value())
	} else {
		var c tea.Cmd
		m.list, c = m.list.Update(msg)
		cmds = append(cmds, c)
	}
	return m, tea.Batch(cmds...)
}

func (m *versionListModel) filter(q string) {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		m.list.SetItems(m.all)
		return
	}
	filtered := []list.Item{}
	for _, it := range m.all {
		vi := it.(versionItem)
		if strings.Contains(strings.ToLower(vi.version), q) || strings.Contains(strings.ToLower(vi.status), q) {
			filtered = append(filtered, it)
		}
	}
	m.list.SetItems(filtered)
}

func (m versionListModel) View() string {
	return m.input.View() + "\n" + m.list.View()
}

func RunVersionList(items []VersionListItem) (string, error) {
	litems := make([]list.Item, 0, len(items))
	for _, it := range items {
		litems = append(litems, versionItem{version: it.Version, status: it.Status})
	}
	p := tea.NewProgram(newVersionListModel(litems))
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	mm := m.(versionListModel)
	if mm.quit {
		return "", nil
	}
	return mm.selected, nil
}

type VersionListItem struct {
	Version string
	Status  string
}

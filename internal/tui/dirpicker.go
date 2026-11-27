package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type dirEntryKind int

const (
	dirUseCurrent dirEntryKind = iota
	dirParent
	dirVolumes
	dirDrive
	dirFolder
)

type dirEntry struct {
	label string
	path  string
	kind  dirEntryKind
}

// DirPicker is a bubbletea model for choosing a directory. It browses into
// folders and, on Windows, lets the user switch between drive volumes.
type DirPicker struct {
	cur      string
	prev     string
	volumes  bool
	entries  []dirEntry
	cursor   int
	selected string
	quit     bool
	width    int
}

func NewDirPicker(start string) DirPicker {
	if start == "" {
		start, _ = os.UserHomeDir()
	}
	start = filepath.Clean(start)
	if info, err := os.Stat(start); err != nil || !info.IsDir() {
		if home, err := os.UserHomeDir(); err == nil {
			start = home
		} else {
			start = string(filepath.Separator)
		}
	}
	m := DirPicker{cur: start, width: 80}
	m.load()
	return m
}

func (m DirPicker) Init() tea.Cmd { return nil }

func (m *DirPicker) load() {
	m.entries = nil
	if m.volumes {
		m.entries = append(m.entries, dirEntry{label: ".. Back to folders", kind: dirParent})
		for _, v := range listVolumes() {
			m.entries = append(m.entries, dirEntry{label: "[" + v + "]", path: v, kind: dirDrive})
		}
	} else {
		m.entries = append(m.entries, dirEntry{label: "* Use this folder", path: m.cur, kind: dirUseCurrent})
		parent := filepath.Dir(m.cur)
		if parent != m.cur {
			m.entries = append(m.entries, dirEntry{label: ".. Parent folder", path: parent, kind: dirParent})
		}
		if runtime.GOOS == "windows" {
			m.entries = append(m.entries, dirEntry{label: "[Volumes: C:, D:, ...]", kind: dirVolumes})
		}
		var names []string
		if items, err := os.ReadDir(m.cur); err == nil {
			for _, e := range items {
				if !e.IsDir() {
					continue
				}
				if e.Name() == "." || e.Name() == ".." {
					continue
				}
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			m.entries = append(m.entries, dirEntry{label: "+ " + name, path: filepath.Join(m.cur, name), kind: dirFolder})
		}
	}
	m.cursor = 0
	for i, e := range m.entries {
		if e.kind == dirFolder || e.kind == dirDrive {
			m.cursor = i
			break
		}
	}
}

func (m DirPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			return m, nil
		case "home":
			m.cursor = 0
			return m, nil
		case "end":
			m.cursor = len(m.entries) - 1
			return m, nil
		case "enter", "right", "l":
			m.activate()
			if m.selected != "" {
				return m, tea.Quit
			}
			return m, nil
		case "left", "h", "backspace":
			m.goBack()
			return m, nil
		case "s":
			m.selected = m.cur
			return m, tea.Quit
		case "v", "V":
			if runtime.GOOS == "windows" {
				if m.volumes {
					m.volumes = false
					m.cur = m.prev
				} else {
					m.prev = m.cur
					m.volumes = true
				}
				m.load()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *DirPicker) activate() {
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return
	}
	e := m.entries[m.cursor]
	switch e.kind {
	case dirUseCurrent:
		// Only cancellation sets quit; selections are returned via selected.
		m.selected = m.cur
	case dirParent:
		if m.volumes {
			m.volumes = false
			m.cur = m.prev
		} else {
			m.cur = filepath.Dir(m.cur)
		}
		m.load()
	case dirVolumes:
		m.prev = m.cur
		m.volumes = true
		m.load()
	case dirDrive, dirFolder:
		m.cur = e.path
		m.volumes = false
		m.load()
	}
}

func (m *DirPicker) goBack() {
	if m.volumes {
		m.volumes = false
		m.cur = m.prev
	} else {
		m.cur = filepath.Dir(m.cur)
	}
	m.load()
}

func (m DirPicker) View() string {
	title := RenderTitle("Move gvm - choose the new folder")
	where := StyleActive.Render("Location: " + m.cur)
	var b strings.Builder
	b.WriteString(title + "\n")
	b.WriteString(where + "\n\n")
	for i, e := range m.entries {
		line := "  " + e.label
		if i == m.cursor {
			line = StyleActive.Render("> " + e.label)
		}
		b.WriteString(line + "\n")
	}
	help := "up/down: move   enter/right: open   left: parent   s: use current   v: volumes   q: cancel"
	if runtime.GOOS != "windows" {
		help = "up/down: move   enter/right: open   left: parent   s: use current   q: cancel"
	}
	b.WriteString("\n" + StyleHint.Render(help))
	return b.String()
}

// RunDirPicker shows the folder browser and returns the chosen directory.
// An empty result means the user cancelled.
func RunDirPicker(start string) (string, error) {
	p := tea.NewProgram(NewDirPicker(start))
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	m := final.(DirPicker)
	if m.quit || m.selected == "" {
		return "", nil
	}
	return m.selected, nil
}

func listVolumes() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	var out []string
	for c := 'A'; c <= 'Z'; c++ {
		root := string(c) + `:\`
		if _, err := os.Stat(root); err != nil {
			continue
		}
		if filepath.VolumeName(root) != "" {
			out = append(out, root)
		}
	}
	return out
}

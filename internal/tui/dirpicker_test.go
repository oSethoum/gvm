package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDirPickerSelectCurrent(t *testing.T) {
	dir := t.TempDir()
	m := NewDirPicker(dir)

	// "Use this folder" is the first entry; Home then Enter must select it.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = next.(DirPicker)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DirPicker)
	if m.selected != filepath.Clean(dir) {
		t.Fatalf("enter did not select current folder: got %q, want %q", m.selected, dir)
	}
	if m.quit {
		t.Fatal("selection must not be treated as a cancel")
	}

	// The s shortcut must also return the folder instead of cancelling.
	m = NewDirPicker(dir)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = next.(DirPicker)
	if m.selected != filepath.Clean(dir) || m.quit {
		t.Fatalf("s shortcut failed: selected=%q quit=%v", m.selected, m.quit)
	}
}

func TestDirPickerCancel(t *testing.T) {
	m := NewDirPicker(t.TempDir())
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(DirPicker)
	if !m.quit || m.selected != "" {
		t.Fatalf("cancel state incorrect: selected=%q quit=%v", m.selected, m.quit)
	}
}

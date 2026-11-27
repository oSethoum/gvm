package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMoveModelKeepsPercentageOnBar(t *testing.T) {
	m := NewMoveModel(`D:\gvm`, nil)
	next, _ := m.Update(MoveProgressMsg{Copied: 50, Total: 100})
	m = next.(MoveModel)
	view := m.View()
	status := strings.SplitN(view, "\n", 2)[0]
	if strings.Contains(status, "%") {
		t.Fatalf("status line should not repeat the percentage: %q", status)
	}
	if !strings.Contains(view, "Moving gvm to") || !strings.Contains(view, "%") {
		t.Fatalf("expected status line and bar percentage: %q", view)
	}
}

func TestMoveModelCancel(t *testing.T) {
	called := false
	m := NewMoveModel(`D:\gvm`, func() { called = true })
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = next.(MoveModel)
	if !called {
		t.Fatal("ctrl+c did not call cancel")
	}
	if cmd != nil {
		t.Fatal("cancel should wait for the copy to clean up before quitting")
	}
	if !strings.Contains(m.View(), "Cancelling") {
		t.Fatalf("missing cancel status: %q", m.View())
	}
	next, _ = m.Update(MoveDoneMsg{Err: context.Canceled})
	m = next.(MoveModel)
	if !strings.Contains(m.View(), "not changed") {
		t.Fatalf("missing cancellation summary: %q", m.View())
	}
}

package tui

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type MoveProgressMsg struct {
	Copied int64
	Total  int64
}
type MoveDoneMsg struct {
	Err error
}

type MoveModel struct {
	progress progress.Model
	status   string
	total    int64
	done     bool
	Err      error
	cancel   func()
}

func NewMoveModel(dst string, cancel func()) MoveModel {
	return MoveModel{
		progress: progress.New(progress.WithDefaultGradient()),
		status:   "Moving gvm to " + dst,
		cancel:   cancel,
	}
}

func (m MoveModel) Init() tea.Cmd { return nil }

func (m MoveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MoveProgressMsg:
		m.total = msg.Total
		if m.total > 0 {
			pct := float64(msg.Copied) / float64(m.total)
			if pct > 1 {
				pct = 1
			}
			return m, m.progress.SetPercent(pct)
		}
		return m, nil
	case MoveDoneMsg:
		m.done = true
		m.Err = msg.Err
		switch {
		case msg.Err == nil:
			m.status = "Move complete"
			return m, tea.Batch(m.progress.SetPercent(1), tea.Quit)
		case errors.Is(msg.Err, context.Canceled):
			m.status = "Move cancelled - gvm was not changed"
		default:
			m.status = "Move failed: " + msg.Err.Error()
		}
		return m, tea.Quit
	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" && m.cancel != nil {
			m.cancel()
			m.status = "Cancelling - cleaning up partial copy..."
		}
	}
	return m, nil
}

func (m MoveModel) View() string {
	if m.done {
		return m.status + "\n"
	}
	if m.total > 0 {
		return fmt.Sprintf("%s\n%s\n", m.status, m.progress.View())
	}
	return m.status + "\n"
}

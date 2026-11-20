package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type DownloadProgressMsg struct {
	Downloaded, Total int64
}
type DownloadDoneMsg struct {
	Err error
}

type DownloadModel struct {
	progress progress.Model
	percent  float64
	status   string
	done     bool
	Err      error
}

func NewDownloadModel() DownloadModel {
	return DownloadModel{
		progress: progress.New(progress.WithDefaultGradient()),
		status:   "Downloading...",
	}
}

func (m DownloadModel) Init() tea.Cmd { return nil }

func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DownloadProgressMsg:
		if msg.Total > 0 {
			m.percent = float64(msg.Downloaded) / float64(msg.Total)
		}
		cmd := m.progress.SetPercent(m.percent)
		return m, cmd
	case DownloadDoneMsg:
		m.done = true
		m.Err = msg.Err
		if msg.Err == nil {
			m.status = "Done"
		} else {
			m.status = "Error: " + msg.Err.Error()
		}
		return m, tea.Quit
	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m DownloadModel) View() string {
	if m.done {
		return m.status + "\n"
	}
	return fmt.Sprintf("%s\n%s\n", m.status, m.progress.View())
}

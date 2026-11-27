package tui

import (
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
	status   string
	total    int64
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
		m.total = msg.Total
		if m.total > 0 {
			pct := float64(msg.Downloaded) / float64(m.total)
			if pct > 1 {
				pct = 1
			}
			return m, m.progress.SetPercent(pct)
		}
		return m, nil
	case DownloadDoneMsg:
		m.done = true
		m.Err = msg.Err
		if msg.Err != nil {
			m.status = "Download failed: " + msg.Err.Error()
		} else {
			m.status = "Download complete"
			return m, tea.Batch(m.progress.SetPercent(1), tea.Quit)
		}
		return m, tea.Quit
	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m DownloadModel) View() string {
	if m.done {
		return m.status + "\n"
	}
	if m.total > 0 {
		return m.status + "\n" + m.progress.View() + "\n"
	}
	return m.status + "\n"
}

package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type spinnerModel struct {
	spinner spinner.Model
	msg     string
	done    bool
	err     error
}

type spinnerDoneMsg struct{ err error }

func newSpinnerModel(msg string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return spinnerModel{spinner: s, msg: msg}
}

func (m spinnerModel) Init() tea.Cmd { return m.spinner.Tick }

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.msg)
}

func RunSpinner(msg string, task func() error) error {
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// non-interactive: run without spinner
		return task()
	}
	p := tea.NewProgram(newSpinnerModel(msg))
	go func() {
		err := task()
		p.Send(spinnerDoneMsg{err: err})
	}()
	m, err := p.Run()
	if err != nil {
		return err
	}
	return m.(spinnerModel).err
}

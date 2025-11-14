package espinner

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The bubbletea.Msg sent when the spinner should stop
type spinnerMsgStop struct {
	err error
}

func (s spinnerMsgStop) Error() string {
	return s.err.Error()
}

type SpinnerTask = func() error

type Spinner = spinner.Spinner

// Spinner style definition
type SpinnerStyle struct {
	ProgressStyle lipgloss.Style
	SuccessStyle  lipgloss.Style
	FailureStyle  lipgloss.Style
}

var SpinnerStyleDefault = SpinnerStyle{
	ProgressStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Faint(true),
	SuccessStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	FailureStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
}

// Bubbletea model of the spinner, wraps spinner.Model and contains the task
// to execute
type SpinnerModel struct {
	title string
	task  SpinnerTask
	inner spinner.Model
	style SpinnerStyle
	err   error
	done  bool
}

// Create a new SpinnerModel.
func NewSpinner(title string, task SpinnerTask) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Line
	return SpinnerModel{
		title: title,
		task:  task,
		style: SpinnerStyleDefault,
		inner: s,
		err:   nil,
		done:  false,
	}
}

// Initialize the SpinnerModel
func (m SpinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.inner.Tick,
		func() tea.Msg {
			err := m.task()
			return spinnerMsgStop{err: err}
		},
	)
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case spinnerMsgStop:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
		}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

func (m SpinnerModel) View() string {
	s := ""
	if !m.done {
		s += m.style.ProgressStyle.Render(fmt.Sprintf("%s %s", m.inner.View(), m.title))
	} else {
		if m.err != nil {
			s += m.style.FailureStyle.Render(fmt.Sprintf("* %s ... Failed: %v", m.title, m.err))
		} else {
			s += m.style.SuccessStyle.Render(fmt.Sprintf("* %s ... Done", m.title))
		}
	}
	s += "\n"
	return s
}

func (m SpinnerModel) Err() error {
	return m.err
}

// Specify the style of the SpinnerModel.
//
//	s := espinner.NewSpinner(...).WithStyle(etable.SpinnerStyleDefault)
func (m SpinnerModel) WithStyle(s SpinnerStyle) SpinnerModel {
	m.style = s
	return m
}

// Specify the spinner of the SpinnerModel.
//
//	s := espinner.NewSpinner(...).WithSpinner(spinner.Dot)
func (m SpinnerModel) WithSpinner(s Spinner) SpinnerModel {
	m.inner.Spinner = s
	return m
}

// Specify the spinner style of the SpinnerModel.
//
//	s := espinner.NewSpinner(...).WithStyle(etable.SpinnerStyleDefault)
func (m SpinnerModel) WithSpinnerStyle(s lipgloss.Style) SpinnerModel {
	m.inner.Style = s
	return m
}

// Run the SpinnerModel.
func (s *SpinnerModel) Spin() error {
	tp := tea.NewProgram(s)
	if _, err := tp.Run(); err != nil {
		return err
	}
	return s.err
}

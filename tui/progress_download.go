package tui

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyleDownload = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

const (
	padding  = 2
	maxWidth = 80
)

type ProgressMsg float64

type ProgressErrMsg struct{ err error }

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type DownloadProgressWriter struct {
	Total      int
	downloaded int
	File       *os.File
	Reader     io.Reader
	OnProgress func(float64)
}

type DownloadProgressModel struct {
	Pw       *DownloadProgressWriter
	Progress progress.Model
	err      error
}

func (m DownloadProgressModel) Init() tea.Cmd {
	return nil
}

func (m DownloadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.Progress.Width = msg.Width - padding*2 - 4
		if m.Progress.Width > maxWidth {
			m.Progress.Width = maxWidth
		}
		return m, nil

	case ProgressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case ProgressMsg:
		var cmds []tea.Cmd

		if msg >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.Progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m DownloadProgressModel) View() string {
	if m.err != nil {
		return "Error downloading: " + m.err.Error() + "\n"
	}

	pad := strings.Repeat(" ", padding)
	return "Downloading...\n" +
		pad + m.Progress.View() + "\n\n" +
		pad + helpStyleDownload("Press any key to quit")
}

func (pw *DownloadProgressWriter) Start(p *tea.Program) {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.File, io.TeeReader(pw.Reader, pw))
	if err != nil {
		p.Send(ProgressErrMsg{err})
	}
}

func (pw *DownloadProgressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.Total > 0 && pw.OnProgress != nil {
		pw.OnProgress(float64(pw.downloaded) / float64(pw.Total))
	}
	return len(p), nil
}

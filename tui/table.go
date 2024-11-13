package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TableModel struct {
	table table.Model
}

func (m TableModel) Init() tea.Cmd { return nil }

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func Table(cols []string, rows [][]string) {
	var columns []table.Column
	for _, col := range cols {
		columns = append(columns, table.Column{
			Title: col,
			Width: 10,
		})
	}

	var finalRows []table.Row

	for _, row := range rows {
		newRow := table.Row{}
		for _, ele := range row {
			newRow = append(newRow, ele)
		}

		finalRows = append(finalRows, newRow)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(finalRows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := TableModel{t}
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running spinner:", err)
		os.Exit(1)
	}

	p.Send(tea.Quit())
}

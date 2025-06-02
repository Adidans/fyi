package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/cpu"
)

type model struct {
	currTime   time.Time
	cpuPercent []float64
	width      int
	height     int
}

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

func initialModel() model {
	return model{
		currTime:   time.Now(),
		cpuPercent: []float64{},
	}
}

func (m model) Init() tea.Cmd {
	return tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tickMsg:
		m.currTime = time.Now()
		percent, err := cpu.Percent(0, false)
		if err != nil {
			m.cpuPercent = []float64{0}
		}
		m.cpuPercent = percent
		return m, tick
	}
	return m, nil
}

func (m model) View() string {
	if len(m.cpuPercent) == 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			"Loading...")
	}
	header := "FYI"
	body := fmt.Sprintf("Current Time: %s\nCPU Usage: %.2f%%", m.currTime.Format(time.DateTime), m.cpuPercent[0])
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, header, body))
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

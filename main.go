package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	currTime time.Time
}

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

func initialModel() model {
	return model{
		currTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tickMsg:
		m.currTime = time.Now()
		return m, tick
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Current time: %s\nPress 'q' to quit.", m.currTime.Format(time.DateTime))
}

func main() {
	p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}

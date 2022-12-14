package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ID string

type guy struct {
	id     ID // immutable (lol)
	name   string
	status string
	online bool
}

type message struct {
	sender  ID
	message string
}

type me struct {
	id     ID
	name   textinput.Model
	status textinput.Model
}

type model struct {
	me          me
	guys        []guy
	messages    []message
	composition textinput.Model
}

func initialModel() model {
	comp := textinput.New()
	comp.Focus()
	myName := textinput.New()
	myName.SetValue("anon")
	return model{
		me: me{
			id:     "0xlolol",
			name:   myName,
			status: textinput.New(),
		},
		guys: []guy{
			guy{
				id:     "1",
				name:   "bob",
				online: true,
			},
			guy{
				id:     "2",
				name:   "fred",
				online: false,
			},
		},
		messages: []message{
			message{
				sender:  "1",
				message: "hi everyone",
			},
			message{
				sender:  "2",
				message: "hi bob",
			},
		},
		composition: comp,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.composition.Focused() {
				m.messages = append(m.messages, message{
					sender:  m.me.id,
					message: m.composition.Value(),
				})
				m.composition.Reset()
			}
		case "tab":
			switch {
			case m.composition.Focused():
				m.composition.Blur()
				m.me.name.Focus()
			case m.me.name.Focused():
				m.me.name.Blur()
				m.me.status.Focus()
			case m.me.status.Focused():
				m.me.status.Blur()
				m.composition.Focus()
			}
		}
	}

	var result [3]tea.Cmd
	m.composition, result[0] = m.composition.Update(msg)
	m.me.name, result[1] = m.me.name.Update(msg)
	m.me.status, result[2] = m.me.status.Update(msg)
	return m, tea.Batch(result[:]...)
}

var (
	onlineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	offlineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func (m model) View() string {
	// me
	var b strings.Builder
	b.WriteString(m.me.name.View())
	b.WriteString(" ")
	b.WriteString(m.me.status.View())
	b.WriteString("\n\n")

	// guys
	for _, guy := range m.guys {
		style := onlineStyle
		if !guy.online {
			style = offlineStyle
		}
		b.WriteString(style.Render(guy.name))
		b.WriteString(" ")
		b.WriteString(guy.status)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// messages
	{
		nameByID := make(map[ID]string)
		for _, guy := range m.guys {
			nameByID[guy.id] = guy.name
		}
		nameByID[m.me.id] = m.me.name.Value()
		msgs := m.messages
		window_len := 8
		if len(msgs) > window_len {
			msgs = msgs[len(msgs)-window_len:]
		}
		for _, msg := range msgs {
			name, ok := nameByID[msg.sender]
			if !ok {
				name = fmt.Sprintf("<user %v>", msg.sender)
			}
			fmt.Fprintf(&b, "%v: %v\n", name, msg.message)
		}
	}
	b.WriteString("\n")

	// composition
	b.WriteString(m.composition.View())

	b.WriteString("\n")
	return b.String()
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

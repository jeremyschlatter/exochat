package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
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

func initialModel(me me) model {
	comp := textinput.New()
	comp.Focus()
	return model{
		me: me,
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
			switch {
			case m.composition.Focused():
				m.messages = append(m.messages, message{
					sender:  m.me.id,
					message: m.composition.Value(),
				})
				m.composition.Reset()
			case m.me.name.Focused(), m.me.status.Focused():
				if err := save_id(m.me); err != nil {
					return m, tea.Quit
				}
				m.me.name.Blur()
				m.me.status.Blur()
				m.composition.Focus()
			}
		case "tab":
			switch {
			case m.composition.Focused():
				m.composition.Blur()
				m.me.name.Focus()
			case m.me.name.Focused():
				m.me.name.Blur()
				if err := save_id(m.me); err != nil {
					return m, tea.Quit
				}
				m.me.status.Focus()
			case m.me.status.Focused():
				m.me.status.Blur()
				if err := save_id(m.me); err != nil {
					return m, tea.Quit
				}
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

func save_filename() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path.Join(d, "exochat"), 0755); err != nil {
		return "", err
	}
	return path.Join(d, "exochat", "id.json"), nil
}

func save_id(m me) error {
	fname, err := save_filename()
	if err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(map[string]string{
		"id":     string(m.id),
		"name":   m.name.Value(),
		"status": m.status.Value(),
	})
}

func load_id() (*me, error) {
	fname, err := save_filename()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var x struct {
		ID     ID
		Name   string
		Status string
	}
	if err := json.NewDecoder(f).Decode(&x); err != nil {
		return nil, err
	}
	result := &me{
		id:     x.ID,
		name:   textinput.New(),
		status: textinput.New(),
	}
	result.name.SetValue(x.Name)
	result.status.SetValue(x.Status)
	return result, nil
}

func main() {
	myself, err := load_id()
	if err != nil {
		if os.IsNotExist(err) {
			name := textinput.New()
			name.SetValue("anon")
			myself = &me{
				id:     ID(uuid.NewString()),
				name:   name,
				status: textinput.New(),
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if _, err := tea.NewProgram(initialModel(*myself)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

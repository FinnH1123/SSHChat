package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maaslalani/gambit/style"
)

type NoteMsg string

type Message struct {
	content string
	sender  string
}

type SharedChat struct {
	user     *User
	messages []Message
	offset   int
	inputBox textinput.Model
	typing   bool
	sync     chan tea.Msg
}

func NewSharedChat(u *User, sync chan tea.Msg) *SharedChat {
	input := textinput.New()
	input.CharLimit = 120
	input.Width = 30
	input.Prompt = ""
	msgs, err := GetRoomMessages(u.room.db, u.room.id)
	if err != nil {
		log.Printf("failed to get msgs: %v\n", err)
	}

	r := &SharedChat{
		user:     u,
		sync:     sync,
		typing:   true,
		offset:   0,
		inputBox: input,
		messages: msgs,
	}
	return r
}

func (r *SharedChat) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, r.inputBox.Focus())
}

func (r *SharedChat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case Message:
		r.messages = append(r.messages, msg)
	case NoteMsg:
		r.messages = append(r.messages, Message{sender: "server", content: string(msg)})
		return r, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			r.user.Close()
		case "up":
			r.offset++
			if r.offset >= len(r.messages)-9 {
				r.offset = len(r.messages) - 9
			}
		case "down":
			r.offset--
			if r.offset < 0 {
				r.offset = 0
			}
		case "enter":
			if r.typing {
				if r.inputBox.Value() != "" {
					r.typing = false
					msg := Message{content: r.inputBox.Value(), sender: r.user.session.User()}
					r.user.room.SendMsg(msg)
					_, err := CreateMessage(r.user.room.db, msg, r.user.room.id)
					if err != nil {
						log.Printf("failed to create msg: %v\n", err)
					}
					r.inputBox.SetValue("")
					r.typing = true
				}
				r.inputBox.Focus()
			} else {
				r.inputBox.Focus()
			}

		}

		if r.typing {
			r.inputBox.Focus()
		}
	}
	if r.typing {
		var (
			cmds []tea.Cmd = make([]tea.Cmd, 1)
		)
		r.inputBox, cmds[0] = r.inputBox.Update(msg)
		return r, tea.Batch(cmds...)
	}
	return r, tea.Batch(cmds...)
}

func (r *SharedChat) View() string {
	s := strings.Builder{}
	s.WriteRune('\n')
	s.WriteString(style.Faint(fmt.Sprintf("In room %s as %s", r.user.room.id, r.user.session.User())))
	s.WriteRune('\n')
	s.WriteRune('\n')
	a := len(r.messages) - r.offset
	b := len(r.messages) - r.offset
	if len(r.messages) < 9 || a < 9 {
		a = 9
		b = 9
	}
	for i := (a - 9); i < b; i++ {
		s.WriteString(fmt.Sprintf("%s: %s\n\n", r.messages[i].sender, r.messages[i].content))
	}
	s.WriteRune('\n')
	s.WriteRune('\n')
	s.WriteString(fmt.Sprintf("Enter your message: %s", r.inputBox.View()))

	return s.String()
}

package server

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
)

type Room struct {
	id     string
	users  map[string]*User
	sync   chan tea.Msg
	done   chan struct{}
	finish chan string
}

func NewRoom(id string, finish chan string) *Room {
	s := make(chan tea.Msg)
	r := &Room{
		id:     id,
		users:  make(map[string]*User, 0),
		sync:   s,
		done:   make(chan struct{}, 1),
		finish: finish,
	}

	return r
}

func (r *Room) Close() {
	log.Printf("closing room %s", r.id)
	r.SendMsg(NoteMsg("Server closing room.\n"))
	for _, p := range r.users {
		p.Close()
	}

	r.done <- struct{}{}
	r.finish <- r.id
	close(r.sync)
	close(r.done)
}

func (r *Room) SendMsg(m tea.Msg) {
	go func() {
		for _, p := range r.users {
			p.Send(m)
		}
	}()
}

func (r *Room) MakeUser(s ssh.Session) *User {
	pl := &User{
		room:    r,
		session: s,
		key:     PublicKey{key: s.PublicKey()},
	}
	m := NewSharedChat(pl, r.sync)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(s),
		tea.WithOutput(s),
	)
	pl.program = p
	pl.chat = m
	return pl
}

func (r *Room) AddUser(s ssh.Session) (*User, error) {
	k := s.PublicKey()
	if k == nil {
		return nil, fmt.Errorf("no public key presented")
	}
	pub := PublicKey{key: k}
	p, ok := r.users[pub.String()]
	if ok {
		return nil, fmt.Errorf("User %s is already in the room (same public key)", p.session.User())
	}
	p = r.MakeUser(s)
	r.users[pub.String()] = p
	r.SendMsg(NoteMsg(fmt.Sprintf("%s joined the room", p.session.User())))
	return p, nil
}

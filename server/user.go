package server

import (
	"fmt"
	"log"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
)

type User struct {
	room    *Room
	session ssh.Session
	program *tea.Program
	chat    *SharedChat
	key     PublicKey
	once    sync.Once
}

func (p *User) Send(m tea.Msg) {
	if p.program != nil {
		p.program.Send(m)
	} else {
		log.Printf("error sending message to user, program is nil")
	}
}

func (p *User) Close() error {
	p.once.Do(func() {
		defer p.room.SendMsg(NoteMsg(fmt.Sprintf("%s has left the room", p.session.User())))
		defer delete(p.room.users, p.key.String())
		if p.program != nil {
			p.program.Kill()
		}
		p.session.Close()
	})
	return nil
}

func (p *User) StartChat() {
	_, wchan, _ := p.session.Pty()
	errc := make(chan error, 1)
	go func() {
		select {
		case err := <-errc:
			log.Printf("error starting program %s", err)
		case w := <-wchan:
			if p.program != nil {
				p.program.Send(tea.WindowSizeMsg{Width: w.Width, Height: w.Height})
			}
		case <-p.session.Context().Done():
			p.Close()
		}
	}()
	defer p.room.SendMsg(NoteMsg(fmt.Sprintf("%v left the room", *p)))

	m, err := p.program.Run()
	if m != nil {
		p.chat = m.(*SharedChat)
	}
	errc <- err
	p.Close()
}

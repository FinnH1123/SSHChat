package server

import (
	"fmt"
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
)

var id string

func chatMiddleware(srv *Server) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		lipgloss.SetColorProfile(termenv.ANSI256)

		return func(s ssh.Session) {
			_, _, active := s.Pty()
			cmds := s.Command()
			fmt.Println(cmds)
			if !active {
				s.Exit(1)
				return
			}
			if len(cmds) == 0 {
				id = "room1"
			} else {
				id = cmds[0]
			}
			room := srv.FindRoom(id)
			if room == nil {
				log.Printf("room %s is created by %s", id, s.User())
				room = srv.NewRoom(id)
			}

			p, err := room.AddUser(s)

			if err != nil {
				s.Write([]byte(err.Error() + "\n"))
				s.Exit(1)
				return
			}
			log.Printf("%s joined room %s [%s]", s.User(), id, s.RemoteAddr())
			p.StartChat()
			log.Printf("%s left room %s [%s]", s.User(), id, s.RemoteAddr())
			p.room.SendMsg(NoteMsg(fmt.Sprintf("%s left room", s.User())))

			sh(s)
		}
	}
}

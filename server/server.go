package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	gossh "golang.org/x/crypto/ssh"

	"github.com/FinnH1123/SSHChat/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
)

type PublicKey struct {
	key ssh.PublicKey
}

func (pk PublicKey) String() string {
	return fmt.Sprint(gossh.MarshalAuthorizedKey(pk.key))
}

type Server struct {
	host  string
	port  int
	db    *sql.DB
	srv   *ssh.Server
	rooms map[string]*Room
}

// NewServer creates a new server.
func NewServer(keyPath, host string, port int) (*Server, error) {
	config.Setup()
	db, err := InitialiseDB()
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	s := &Server{
		host:  host,
		port:  port,
		rooms: make(map[string]*Room),
		db:    db,
	}
	ws, err := wish.NewServer(
		wish.WithPublicKeyAuth(publicKeyHandler),
		wish.WithHostKeyPath(keyPath),
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			chatMiddleware(s),
		),
	)
	if err != nil {
		return nil, err
	}
	s.srv = ws
	return s, nil
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// Shutdown shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	for _, room := range s.rooms {
		room.Close()
	}
	return s.srv.Shutdown(ctx)
}

func publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func (s *Server) FindRoom(id string) *Room {
	r, ok := s.rooms[id]
	if !ok {
		return nil
	}
	return r
}

func (s *Server) NewRoom(id string) *Room {
	finish := make(chan string, 1)
	go func() {
		id := <-finish
		log.Printf("deleting room %s", id)
		delete(s.rooms, id)
		close(finish)
	}()

	room := NewRoom(id, finish, s.db, s)
	s.rooms[id] = room
	return room
}

func (s *Server) SwitchRoom(newId string, user *User) error {
	room := s.FindRoom(newId)
	if room == nil {
		room = s.NewRoom(newId)
	}
	if room.users[user.key.String()] != nil {
		return fmt.Errorf("%s", "publickey already in that room")
	}
	user.room.SendMsg(NoteMsg(fmt.Sprintf("%s has left the room", user.session.User())))
	delete(user.room.users, user.key.String())
	user.program.Kill()
	user = room.MakeUser(user.session)
	room.users[user.key.String()] = user
	user.StartChat()
	user.Send(tea.KeyMsg{Type: 1, Runes: []int32{2}, Alt: true})
	return nil
}

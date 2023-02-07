package server

import (
	"context"
	"fmt"
	"log"

	gossh "golang.org/x/crypto/ssh"

	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
)

type PublicKey struct {
	key ssh.PublicKey
}

func (pk PublicKey) String() string {
	return fmt.Sprintf("%s", gossh.MarshalAuthorizedKey(pk.key))
}

type Server struct {
	host  string
	port  int
	srv   *ssh.Server
	rooms map[string]*Room
}

// NewServer creates a new server.
func NewServer(keyPath, host string, port int) (*Server, error) {
	s := &Server{
		host:  host,
		port:  port,
		rooms: make(map[string]*Room),
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

func passwordHandler(ctx ssh.Context, password string) bool {
	return true
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

	room := NewRoom(id, finish)
	s.rooms[id] = room
	return room
}

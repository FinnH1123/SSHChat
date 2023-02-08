package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FinnH1123/SSHChat/server"
)

var (
	host string = "0.0.0.0"
	port int    = 23234
	key  string = ""
)

func main() {

	s, err := server.NewServer(key, host, port)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting ssh server on %s:%d", host, port)
	go func() {
		if err = s.Start(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Print("Stopping ssh server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

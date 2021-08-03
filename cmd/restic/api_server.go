package main

import (
	. "fmt"

	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
)

var logger = httplog.NewLogger("restic-api", httplog.Options{
	Concise: true,
})

type Server struct {
	*http.Server
	// Implementing graceful shutdown
	shutdown       chan bool
	done           chan bool
	shutdownAction func()
}

type Config struct {
	Port string
}

func New(config Config, createRouter func(server *Server) http.Handler) *Server {

	httpServer := &http.Server{
		Addr: "0.0.0.0:" + config.Port,
	}

	s := &Server{httpServer, make(chan bool), make(chan bool), nil}

	go func() {
		<-s.shutdown

		if err := s.Shutdown(context.Background()); err != nil {
			Warnf("shutdown error: %v", err)
		}

		if s.shutdownAction != nil {
			s.shutdownAction()
		}

		s.done <- true
	}()

	httpServer.Handler = createRouter(s)

	return s
}

func (s *Server) GracefulShutdown() {
	s.shutdown <- true
}

func (s *Server) SetGracefulShutdownAction(action func()) {
	s.shutdownAction = action
}

func (s *Server) ListenAndServe() error {
	err := s.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}

	<-s.done
	return err
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	for header, value := range r.Header {
		Verbosef("%v : %v\n", header, value)
	}
	Verbosef("RemoteAddr : %v\n", r.RemoteAddr)
	Fprintf(w, "Welcome\n")
}

func (s *Server) hey(w http.ResponseWriter, r *http.Request) {
	Fprintf(w, "Hey\n")
}

func (s *Server) slow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	Fprintf(w, "Done slow operation.\n")
}

func (s *Server) exit(w http.ResponseWriter, r *http.Request) {
	s.GracefulShutdown()
	Fprintf(w, "Exiting...\n")
}

func startServer(port string) error {

	config := Config{port}

	createRouter := func(server *Server) http.Handler {
		router := chi.NewRouter()

		router.Use(httplog.RequestLogger(logger))
		router.Use(OnlyLocal)

		router.Get("/", server.home)
		router.Get("/hi", server.hey)
		router.Get("/slow", server.slow)
		router.Get("/exit", server.exit)

		return router
	}

	server := New(config, createRouter)

	Verbosef("Server is running:\n    ==> localhost:%v\n", port)

	server.SetGracefulShutdownAction(func() {
		Verbosef("We are done shutting down\n")
	})

	err := server.ListenAndServe()
	if err != nil {
		Warnf("server error: %v\n", err)
	}

	return err
}

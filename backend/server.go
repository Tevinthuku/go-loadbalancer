package backend

import (
	"fmt"
	"html"
	"net/http"
)

type Server struct {
	address string
}

func NewServer(address string) *Server {
	return &Server{address: address}
}

func (s *Server) Start() error {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// wild-card route for all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	return http.ListenAndServe(s.address, nil)
}

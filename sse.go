package sse

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	mu    sync.RWMutex
	chans map[uuid.UUID]chan message
}

func New() *Server {
	return &Server{
		chans: make(map[uuid.UUID]chan message),
	}
}

func (s *Server) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	chanId := uuid.New()
	c := make(chan message)
	s.mu.Lock()
	s.chans[chanId] = c
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.chans, chanId)
		close(c)
		s.mu.Unlock()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("ResponseWriter does not implement http.Flusher")
		return
	}
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			fmt.Println("SSE " + chanId.String() + " closed")
			return
		case m := <-c:
			if err := m.writeTo(w); err != nil {
				log.Println("Unable to write")
				return
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				log.Println("Unable to write")
				return
			}
			flusher.Flush()
		}

	}
}

func (s *Server) Send(id int, event string, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m := message{
		id:    id,
		event: event,
		data:  data,
	}

	for _, c := range s.chans {
		c <- m
	}
}

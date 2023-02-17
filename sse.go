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
	chans map[uuid.UUID]chan Message
}

func New() *Server {
	return &Server{
		chans: make(map[uuid.UUID]chan Message),
	}
}

type Message struct {
	Id    int    `json:"id,omitempty"`
	Event string `json:"event,omitempty"`
	Data  []byte `json:"data,omitempty"`
}

func (s *Server) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	chanId := uuid.New()
	c := make(chan Message)
	s.mu.Lock()
	s.chans[chanId] = c
	s.mu.Unlock()
	defer func() {
		delete(s.chans, chanId)
		close(c)
	}()
	fmt.Println("SSE " + chanId.String() + " opened")

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
			if m.Id != 0 {
				if _, err := w.Write([]byte(fmt.Sprintf("id: %d\n", m.Id))); err != nil {
					log.Println("Unable to write")
					return
				}
			}
			if m.Event != "" {
				if _, err := w.Write([]byte(fmt.Sprintf("event: %s\n", m.Event))); err != nil {
					log.Println("Unable to write")
					return
				}
			}
			if len(m.Data) != 0 {
				if _, err := w.Write([]byte(fmt.Sprintf("data: %s\n", m.Data))); err != nil {
					log.Println("Unable to write")
					return
				}
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				log.Println("Unable to write")
				return
			}
			flusher.Flush()
		}

	}
}

func (s *Server) Write(m Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.chans {
		c <- m
	}
}

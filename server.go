/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

import (
	"sync"
	"time"
)

const (
	// DefaultBufferSize size of the queue that holds the streams messages.
	DefaultBufferSize = 1024
	// DefaultMaxInactivity of a stream
	DefaultMaxInactivity = time.Second * 5
)

// Server Is our main struct
type Server struct {
	// Specifies the size of the message buffer for each stream
	BufferSize int
	// Enables creation of a stream when a client connects
	AutoStream bool
	Streams    map[string]*Stream
	mu         sync.Mutex
}

// New will create a server and setup defaults
func New() *Server {
	return &Server{
		BufferSize: DefaultBufferSize,
		AutoStream: false,
		Streams:    make(map[string]*Stream),
	}
}

// Close shuts down the server, closes all of the streams and connections
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.Streams {
		s.Streams[id].quit <- true
		delete(s.Streams, id)
	}
}

// GetStream returns a stream by id
func (s *Server) GetStream(id string) *Stream {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Streams[id]
}

// CreateStream will create a new stream and register it
func (s *Server) CreateStream(id string) *Stream {
	// Register new stream
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Streams[id] != nil {
		return s.Streams[id]
	}

	s.Streams[id] = newStream(s.BufferSize)

	return s.Streams[id]
}

// RemoveStream will remove a stream
func (s *Server) RemoveStream(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Streams[id] != nil {
		s.Streams[id].close()
		delete(s.Streams, id)
	}
}

// StreamExists checks whether a stream by a given id exists
func (s *Server) StreamExists(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Streams[id] != nil
}

// Publish sends a mesage to every client in a streamID
func (s *Server) Publish(id string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Streams[id] != nil {
		s.Streams[id].event <- &Event{Data: data}
	}
}

// Register a subscriber
func (s *Server) Register(id string, sub *Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Streams[id].addSubscriber(sub)
}

// GetSubscriber will get an existing subscriber
func (s *Server) GetSubscriber(id string) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, stream := range s.Streams {
		sub := stream.getSubscriber(id)
		if sub != nil {
			return sub
		}
	}

	return nil
}

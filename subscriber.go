/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

import "sync"

// Subscriber ...
type Subscriber struct {
	id          string
	eventid     string
	quit        chan *Subscriber
	connections []chan *Event
	mu          sync.Mutex
}

// NewSubscriber creates a new subscriber with defaults
func NewSubscriber(id string) *Subscriber {
	return &Subscriber{
		id:          id,
		eventid:     "0",
		connections: make([]chan *Event, 0),
	}
}

// Broadcast an event to all of a subscribers connections
func (s *Subscriber) Broadcast(e *Event) {
	s.mu.Lock()
	defer s.mu.Lock()

	for i := range s.connections {
		s.connections[i] <- e
	}
}

// Connect creates a new connection channel on a subscriber
func (s *Subscriber) Connect() chan *Event {
	s.mu.Lock()
	defer s.mu.Lock()

	c := make(chan *Event, 64)
	s.connections = append(s.connections, c)
	return c
}

// Disconnect a subscriber connection from the subscriber
func (s *Subscriber) Disconnect(c chan *Event) {
	for i := range s.connections {
		if s.connections[i] == c {
			close(c)
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
		}
	}

	if len(s.connections) == 0 {
		s.Close()
	}
}

// DisconnectAll closes all subscriber connections
func (s *Subscriber) DisconnectAll() {
	for i := range s.connections {
		close(s.connections[i])
		s.connections = append(s.connections[:i], s.connections[i+1:]...)
	}
	s.Close()
}

// Close will let the stream know that the clients connection has terminated
func (s *Subscriber) Close() {
	if s.quit != nil {
		s.quit <- s
	}
}

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

import (
	"sync"
)

// Subscriber ...
type Subscriber struct {
	id          string
	quit        chan *Subscriber
	replay      chan *Connection
	connections []*Connection
	mu          sync.Mutex
}

// NewSubscriber creates a new subscriber with defaults
func NewSubscriber(id string) *Subscriber {
	return &Subscriber{
		id:          id,
		connections: make([]*Connection, 0),
	}
}

// Broadcast an event to all of a subscribers connections
func (s *Subscriber) Broadcast(e *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.connections {
		s.connections[i].Send(e)
	}
}

// Connect creates a new connection channel on a subscriber
func (s *Subscriber) Connect() chan *Event {
	return s.ConnectAtID("0")
}

// ConnectAtID creates a new connection and replays events from a given event id
func (s *Subscriber) ConnectAtID(id string) chan *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := Connection{
		conn:    make(chan *Event, 64),
		eventid: id,
	}

	s.connections = append(s.connections, &c)

	go func() {
		s.replay <- &c
	}()

	return c.conn
}

// Disconnect a subscriber connection from the subscriber
func (s *Subscriber) Disconnect(c chan *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.connections) - 1; i >= 0; i-- {
		if s.connections[i].conn == c {
			close(c)
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
		}
	}
}

// DisconnectAll closes all subscriber connections
func (s *Subscriber) DisconnectAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.connections) - 1; i >= 0; i-- {
		if s.connections[i] != nil {
			close(s.connections[i].conn)
		}
		s.connections = append(s.connections[:i], s.connections[i+1:]...)
	}
}

// HasConnections returns true if there are any subscriber connections
func (s *Subscriber) HasConnections() bool {
	return len(s.connections) > 0
}

// Close will let the stream know that the clients connection has terminated
func (s *Subscriber) Close() {
	if s.quit != nil {
		s.quit <- s
	}
}

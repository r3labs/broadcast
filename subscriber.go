/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

// Subscriber ...
type Subscriber struct {
	id         string
	eventid    string
	quit       chan *Subscriber
	connection chan *Event
}

// NewSubscriber creates a new subscriber with defaults
func NewSubscriber(id string) *Subscriber {
	return &Subscriber{
		id:         id,
		eventid:    "0",
		connection: make(chan *Event, 64),
	}
}

// Close will let the stream know that the clients connection has terminated
func (s *Subscriber) close() {
	s.quit <- s
}

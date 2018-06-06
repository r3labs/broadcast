/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

// Stream ...
type Stream struct {
	// Enables replaying of eventlog to newly added subscribers
	AutoReplay  bool
	Eventlog    EventLog
	stats       chan chan int
	subscribers []*Subscriber
	register    chan *Subscriber
	deregister  chan *Subscriber
	replay      chan *Connection
	event       chan *Event
	quit        chan bool
}

// StreamRegistration ...
type StreamRegistration struct {
	id     string
	stream *Stream
}

// newStream returns a new stream
func newStream(bufsize int) *Stream {
	s := &Stream{
		AutoReplay:  true,
		Eventlog:    make(EventLog, 0),
		subscribers: make([]*Subscriber, 0),
		register:    make(chan *Subscriber),
		deregister:  make(chan *Subscriber),
		replay:      make(chan *Connection),
		event:       make(chan *Event, bufsize),
		quit:        make(chan bool),
	}

	s.run()

	return s
}

func (str *Stream) run() {
	go func(str *Stream) {
		for {
			select {
			// Add new subscriber
			case subscriber := <-str.register:
				if str.AutoReplay {
					subscriber.replay = str.replay
				}
				str.subscribers = append(str.subscribers, subscriber)

			// Remove closed subscriber
			case subscriber := <-str.deregister:
				i := str.getSubscriberIndex(subscriber)
				if i != -1 {
					str.removeSubscriber(i)
				}

			// Publish event to subscribers
			case event := <-str.event:
				str.Eventlog.Add(event)
				for i := range str.subscribers {
					str.subscribers[i].Broadcast(event)
				}

			case conn := <-str.replay:
				str.Eventlog.Replay(conn)

			// Shutdown if the server closes
			case <-str.quit:
				// remove connections
				str.removeAllSubscribers()
				str.cleanup()
				return
			}
		}
	}(str)
}

func (str *Stream) close() {
	str.quit <- true
}

func (str *Stream) cleanup() {
	close(str.event)
	close(str.register)
	close(str.deregister)
	close(str.quit)
}

func (str *Stream) getSubscriber(id string) *Subscriber {
	for i := range str.subscribers {
		if str.subscribers[i].id == id {
			return str.subscribers[i]
		}
	}

	return nil
}

func (str *Stream) getSubscriberIndex(sub *Subscriber) int {
	for i := range str.subscribers {
		if str.subscribers[i].id == sub.id {
			return i
		}
	}
	return -1
}

// addSubscriber will register a subscriber on a stream
func (str *Stream) addSubscriber(sub *Subscriber) {
	sub.quit = str.deregister
	sub.replay = str.replay
	str.register <- sub
}

func (str *Stream) removeSubscriber(i int) {
	str.subscribers[i].DisconnectAll()
	str.subscribers = append(str.subscribers[:i], str.subscribers[i+1:]...)
}

func (str *Stream) removeAllSubscribers() {
	for i := range str.subscribers {
		str.subscribers[i].DisconnectAll()
	}

	str.subscribers = str.subscribers[:0]
}

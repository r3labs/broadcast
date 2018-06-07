/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStreamRegisterSubscriber(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	sub := NewSubscriber("test")
	s.addSubscriber(sub)

	assert.Len(t, s.subscribers, 1)
}

func TestStreamDeregisterSubscriber(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	sub := NewSubscriber("test")
	s.addSubscriber(sub)

	s.removeSubscriber(0)

	assert.Len(t, s.subscribers, 0)
}

func TestStreamPublish(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	sub1 := NewSubscriber("test-1")
	sub2 := NewSubscriber("test-2")
	sub3 := NewSubscriber("test-3")

	s.addSubscriber(sub1)
	s.addSubscriber(sub2)
	s.addSubscriber(sub3)

	c1 := sub1.Connect()
	c2 := sub2.Connect()
	c3 := sub3.Connect()

	s.event <- &Event{Data: []byte("ping")}

	assert.Len(t, s.subscribers, 3)

	for i := 0; i < 3; i++ {
		select {
		case event := <-c1:
			assert.Equal(t, event.Data, []byte("ping"))
		case event := <-c2:
			assert.Equal(t, event.Data, []byte("ping"))
		case event := <-c3:
			assert.Equal(t, event.Data, []byte("ping"))
		case <-time.After(time.Second):
			t.Fail()
		}
	}
}

func TestStreamClose(t *testing.T) {
	s := newStream(DefaultBufferSize)

	sub1 := NewSubscriber("test-1")
	sub2 := NewSubscriber("test-2")

	s.addSubscriber(sub1)
	s.addSubscriber(sub2)

	time.Sleep(time.Millisecond * 100)

	assert.Len(t, s.subscribers, 2)

	sub1.Connect()
	sub2.Connect()

	s.close()

	time.Sleep(time.Millisecond * 100)

	assert.Len(t, s.subscribers, 0)
}

func TestStreamNewSubscriberConnect(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	sub := NewSubscriber("test")
	s.addSubscriber(sub)

	c := sub.Connect()

	s.event <- &Event{Data: []byte("ping")}

	select {
	case event := <-c:
		assert.Equal(t, event.Data, []byte("ping"))
	case <-time.After(time.Second):
		t.Fail()
	}
}

func TestStreamExistingSubscriberConnect(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	esub := NewSubscriber("test")
	s.addSubscriber(esub)

	ec := esub.Connect()

	s.event <- &Event{Data: []byte("ping")}

	sub := s.getSubscriber("test")

	assert.NotNil(t, sub)

	c := sub.Connect()

	for i := 0; i < 2; i++ {
		select {
		case event := <-ec:
			assert.Equal(t, event.Data, []byte("ping"))
		case event := <-c:
			assert.Equal(t, event.Data, []byte("ping"))
		case <-time.After(time.Second):
			t.Fail()
		}
	}
}

func TestStreamReplay(t *testing.T) {
	s := newStream(DefaultBufferSize)
	defer s.close()

	for i := 0; i < 10; i++ {
		s.event <- &Event{Data: []byte(strconv.Itoa(i))}
	}

	sub := NewSubscriber("test")
	s.addSubscriber(sub)
	c := sub.Connect()

	for i := 0; i < 10; i++ {
		e := <-c
		assert.Equal(t, string(e.Data), strconv.Itoa(i))
	}
}

func TestStreamInactivity(t *testing.T) {
	s := newStream(DefaultBufferSize)

	s.MaxInactivity = time.Second

	for i := 0; i < 10; i++ {
		s.event <- &Event{Data: []byte(strconv.Itoa(i))}
	}

	sub := NewSubscriber("test")
	s.addSubscriber(sub)
	c := sub.Connect()

	for i := 0; i < 10; i++ {
		e := <-c
		assert.Equal(t, string(e.Data), strconv.Itoa(i))
	}

	sub.DisconnectAll()

	time.Sleep(time.Second * 2)

	assert.True(t, s.closed)
}

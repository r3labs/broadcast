/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package broadcast

// Connection ..
type Connection struct {
	conn    chan *Event
	eventid string
}

// Send an event to a given subscriber connection
func (c *Connection) Send(e *Event) {
	if c.conn != nil {
		c.conn <- e
	}
}

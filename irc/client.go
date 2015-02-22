// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package irc

import (
	"fmt"
	"net"
)

type Hook func(*Client, Message) error

type Client struct {
	conn  *net.Conn
	hooks map[string][]Hook
}

func NewClient(conn *net.Conn) *Client {
	return &Client{
		conn:  conn,
		hooks: make(map[string][]Hook),
	}
}

func (c *Client) Write(format string, argv ...interface{}) error {
	_, err := fmt.Fprintf(*c.conn, "%s\r\n", fmt.Sprintf(format, argv...))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Handle(data string) <-chan error {
	out := make(chan error)
	msg := parseMessage(data)

	hooks, ok := c.hooks[msg.Command]
	if ok {
		for _, hook := range hooks {
			go func() {
				if err := hook(c, msg); err != nil {
					out <- err
				}
			}()
		}
	}

	close(out)
	return out
}

func (c *Client) CmdHook(cmd string, hook Hook) {
	c.hooks[cmd] = append(c.hooks[cmd], hook)
}

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
	"html"
	"net"
	"strings"
	"unicode"
)

type Hook func(*Client, Message) error

type Client struct {
	conn     net.Conn
	hooks    map[string][]Hook
	Nickname string
	Realname string
	Channels []string
}

func NewClient(conn net.Conn) *Client {
	c := &Client{
		conn:  conn,
		hooks: make(map[string][]Hook),
	}

	c.CmdHook("join", joinCmd)
	c.CmdHook("part", partCmd)
	c.CmdHook("kick", kickCmd)

	c.CmdHook("ping", pingCmd)
	return c
}

func (c *Client) Setup(nick, name, host string) {
	c.Nickname = nick
	c.Realname = name

	c.Write("USER %s %s * :%s", c.Nickname, host, c.Realname)
	c.Write("NICK %s", c.Nickname)
}

func (c *Client) Connected(channel string) bool {
	for _, c := range c.Channels {
		if c == channel {
			return true
		}
	}

	return false
}

func (c *Client) Write(format string, argv ...interface{}) error {
	_, err := fmt.Fprintf(c.conn, "%s\r\n", fmt.Sprintf(sanitize(format), argv...))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Handle(data string, ch chan error) {
	msg := parseMessage(data)
	hooks, ok := c.hooks[msg.Command]
	if ok {
		for _, hook := range hooks {
			go func(h Hook) {
				if err := h(c, msg); err != nil {
					ch <- err
				}
			}(hook)
		}
	}
}

func (c *Client) CmdHook(cmd string, hook Hook) {
	c.hooks[cmd] = append(c.hooks[cmd], hook)
}

func joinCmd(client *Client, msg Message) error {
	if msg.Sender.Name == client.Nickname {
		client.Channels = append(client.Channels, msg.Data)
	}

	return nil
}

func partCmd(client *Client, msg Message) error {
	if msg.Sender.Name == client.Nickname {
		client.Channels = remove(msg.Data, client.Channels)
	}

	return nil
}

func kickCmd(client *Client, msg Message) error {
	if msg.Data == client.Nickname {
		target := strings.Fields(msg.Receiver)[0]
		client.Channels = remove(target, client.Channels)
	}

	return nil
}

func pingCmd(client *Client, msg Message) error {
	return client.Write("PONG %s", msg.Data)
}

// sanitize removes all non-printable characters from
// the given string by returning a new string without them.
func sanitize(text string) string {
	mfunc := func(r rune) rune {
		switch {
		case !unicode.IsPrint(r):
			return ' '
		case unicode.IsSpace(r):
			return ' '
		default:
			return r
		}
	}

	escaped := strings.Map(mfunc, html.UnescapeString(text))
	return strings.Join(strings.Fields(escaped), " ")
}

// remove deletes a given element from a given slice. A new slice
// which does not contain the given element is returned.
func remove(element string, slice []string) []string {
	var newSlice []string
	for _, e := range slice {
		if e != element {
			newSlice = append(newSlice, e)
		}
	}

	return newSlice
}

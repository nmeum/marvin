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
	"strings"
)

type Sender struct {
	Name string
	Host string
}

type Message struct {
	Sender   Sender
	Receiver string
	Command  string
	Data     string
}

func parseMessage(line string) Message {
	msg := Message{}
	if len(strings.Fields(line)) < 2 {
		return msg
	}

	if strings.HasPrefix(line, ":") {
		idx := strings.Index(line, " ")
		msg.Sender = Sender{Name: line[1:idx]}

		user := strings.Split(msg.Sender.Name, "!")
		if len(user) >= 2 {
			msg.Sender.Name = user[0]
			msg.Sender.Host = user[1]
		}

		line = line[idx+1:]
	}

	idx := strings.Index(line, " ")
	msg.Command = strings.ToLower(line[:idx])
	line = line[idx+1:]

	if strings.Contains(line, " ") {
		idx = strings.Index(line, ":")
		if idx >= 0 {
			msg.Receiver = strings.TrimSpace(line[0:idx])
			msg.Data = line[idx+1:]
		}
	} else {
		msg.Data = line
	}

	msg.Data = strings.TrimSpace(msg.Data)
	return msg
}

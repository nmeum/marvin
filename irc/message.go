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
	if strings.HasPrefix(line, ":") {
		idx := strings.Index(line, " ")
		msg.Sender = Sender{Name: line[1:idx]}

		if strings.Contains(msg.Sender.Name, "!") {
			s := strings.Split(msg.Sender.Name, "!")
			msg.Sender.Name = s[0]
			msg.Sender.Host = s[1]
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
			msg.Data = strings.TrimSpace(line[idx+1:])
		}
	} else {
		msg.Data = line[1:]
	}

	return msg
}

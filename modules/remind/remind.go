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

package remind

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"strings"
	"time"
)

type Module struct {
	TimeLimit int `json:"time_limit"`
	UserLimit int `json:"user_limit"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "remind"
}

func (m *Module) Help() string {
	return "USAGE: !remind DURATION MSG"
}

func (m *Module) Defaults() {
	m.TimeLimit = 10
	m.UserLimit = 3
}

func (m *Module) Load(client *irc.Client) error {
	users := make(map[string]int)
	client.CmdHook("privmsg", func(c *irc.Client, msg irc.Message) error {
		splited := strings.Fields(msg.Data)
		if len(splited) < 3 || splited[0] != "!remind" {
			return nil
		}

		duration, err := time.ParseDuration(splited[1])
		if err != nil {
			return err
		}

		limit := time.Duration(m.TimeLimit) * time.Hour
		if duration > limit {
			return c.Write("NOTICE %s :%v hours exceeds the limit of %v hours",
				msg.Receiver, duration.Hours(), limit.Hours())
		}

		if users[msg.Sender.Host] >= m.UserLimit {
			return c.Write("NOTICE %s :You can only run %d reminders at a time",
				msg.Receiver, m.UserLimit)
		}

		users[msg.Sender.Host]++
		reminder := strings.Join(splited[2:], " ")
		time.AfterFunc(duration, func() {
			users[msg.Sender.Host]--
			c.Write("PRIVMSG %s :Reminder: %s",
				msg.Sender.Name, reminder)
		})

		return c.Write("NOTICE %s :Reminder setup for %s",
			msg.Receiver, duration.String())
	})

	return nil
}

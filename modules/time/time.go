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

package time

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"time"
)

type Module struct {
	startup time.Time
	Format  string `json:"format"`
	Uptime  bool   `json:"uptime"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "time"
}

func (m *Module) Help() string {
	return "USAGE: !time || !uptime"
}

func (m *Module) Defaults() {
	m.Format = time.RFC1123
	m.Uptime = true
}

func (m *Module) Load(client *irc.Client) error {
	if m.Uptime {
		m.startup = time.Now()
		client.CmdHook("privmsg", m.uptimeCmd)
	}

	client.CmdHook("privmsg", m.dateCmd)
	return nil
}

func (m *Module) dateCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!time" {
		return nil
	}

	return client.Write("NOTICE %s :%s",
		msg.Receiver, time.Now().Format(m.Format))
}

func (m *Module) uptimeCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!uptime" {
		return nil
	}

	duration := time.Now().Sub(m.startup)
	return client.Write("NOTICE %s :%s",
		msg.Receiver, duration.String())
}

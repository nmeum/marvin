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
	Format string `json:"format"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "time"
}

func (m *Module) Help() string {
	return "USAGE: !time"
}

func (m *Module) Defaults() {
	m.Format = time.RFC1123
}

func (m *Module) Load(client *irc.Client) error {
	client.CmdHook("privmsg", m.timeCmd)
	return nil
}

func (m *Module) timeCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!time" {
		return nil
	}

	now := time.Now().UTC()
	return client.Write("NOTICE %s :%s",
		msg.Receiver, now.Format(m.Format))
}

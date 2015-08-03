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

package rejoin

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"strings"
	"time"
)

type Module struct {
	Timeout string `json:"timeout"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "rejoin"
}

func (m *Module) Help() string {
	return "Enables automatic reconnection on kick."
}

func (m *Module) Defaults() {
	m.Timeout = "0m3s"
}

func (m *Module) Load(client *irc.Client) error {
	duration, err := time.ParseDuration(m.Timeout)
	if err != nil {
		return err
	}

	client.CmdHook("kick", func(c *irc.Client, msg irc.Message) error {
		if msg.Data != client.Nickname {
			return nil
		}

		time.Sleep(duration)
		return c.Write("JOIN %s", strings.Fields(msg.Receiver)[0])
	})

	return nil
}

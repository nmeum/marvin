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

package nickserv

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"strings"
)

type Module struct {
	NickServ string `json:"nickserv"`
	Password string `json:"password"`
	Keyword  string `json:"keyword"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "NickServ"
}

func (m *Module) Help() string {
	return "Enables authentication with NickServ."
}

func (m *Module) Defaults() {
	m.NickServ = "NickServ"
	m.Password = ""
	m.Keyword = "identify"
}

func (m *Module) Load(client *irc.Client) error {
	if len(m.Password) <= 0 {
		return nil
	}

	client.CmdHook("notice", func(c *irc.Client, msg irc.Message) error {
		if msg.Sender.Name != m.NickServ || !strings.Contains(msg.Data, m.Keyword) {
			return nil
		}

		return c.Write("PRIVMSG %s :identify %s",
			m.NickServ, m.Password)
	})

	return nil
}

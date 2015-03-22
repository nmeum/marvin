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

package modules

import (
	"encoding/json"
	"fmt"
	"github.com/nmeum/marvin/irc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Module interface {
	Name() string
	Help() string
	Load(*irc.Client) error
	Defaults()
}

type ModuleSet struct {
	client  *irc.Client
	modules []Module
	configs string
}

func NewModuleSet(client *irc.Client, configs string) *ModuleSet {
	return &ModuleSet{client: client, configs: configs}
}

func (m *ModuleSet) Register(module Module) {
	m.modules = append(m.modules, module)
}

func (m *ModuleSet) LoadAll() error {
	if err := os.MkdirAll(m.configs, 0755); err != nil {
		return err
	}

	for _, module := range m.modules {
		fn := fmt.Sprintf("%s.json", module.Name())
		fp := filepath.Join(m.configs, fn)

		module.Defaults()
		data, err := ioutil.ReadFile(fp)
		if err == nil {
			if err := json.Unmarshal(data, &module); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}

		if err := module.Load(m.client); err != nil {
			return err
		}
	}

	m.client.CmdHook("privmsg", m.helpCmd)
	m.client.CmdHook("privmsg", m.moduleCmd)

	return nil
}

func (m *ModuleSet) findModule(name string) Module {
	for _, module := range m.modules {
		if module.Name() == name {
			return module
		}
	}

	return nil
}

func (m *ModuleSet) helpCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!help" && len(m.modules) >= 1 {
		return nil
	}

	var names []string
	for _, module := range m.modules {
		names = append(names, module.Name())
	}

	help := fmt.Sprintf("The following modules are available: %s",
		strings.Join(names, ", "))

	return client.Write("NOTICE %s :%s", msg.Receiver, help)
}

func (m *ModuleSet) moduleCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!help" {
		return nil
	}

	name := strings.ToLower(splited[1])
	module := m.findModule(name)
	if module == nil {
		return client.Write("NOTICE %s :Module %q isn't installed",
			msg.Receiver, name)
	}

	return client.Write("NOTICE %s :%s: %s",
		msg.Receiver, module.Name(), module.Help())
}

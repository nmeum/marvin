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

package veto

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"time"
)

type Module struct {
	timer       *time.Timer
	duration    time.Duration
	DurationStr string `json:"duration"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "veto"
}

func (m *Module) Help() string {
	return "USAGE: !veto"
}

func (m *Module) Defaults() {
	m.DurationStr = "0h90s"
}

func (m *Module) Load(client *irc.Client) (err error) {
	m.duration, err = time.ParseDuration(m.DurationStr)
	if err != nil {
		return
	}

	client.CmdHook("privmsg", m.vetoCmd)
	return
}

func (m *Module) Start() bool {
	ret := false
	m.timer = time.AfterFunc(m.duration, func() {
		ret = true
		m.timer = nil
	})

	return ret
}

func (m *Module) vetoCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!veto" && m.timer != nil {
		return nil
	}

	m.timer.Stop()
	m.timer = nil

	return client.Write("NOTICE %s :%s has invoked his right to veto.",
		msg.Receiver, msg.Sender.Name)
}

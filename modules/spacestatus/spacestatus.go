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

package spacestatus

import (
	"encoding/json"
	"errors"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"io/ioutil"
	"net/http"
	"time"
)

// Supported SpaceAPI version
const apiVersion = "0.13"

type spaceapi struct {
	API   string `json:"api"`
	Space string `json:"space"`
	State struct {
		Open bool `json:"open"`
	} `json:"state"`
}

type Module struct {
	api      *spaceapi
	URL      string `json:"url"`
	Notify   bool   `json:"notify"`
	Interval string `json:"interval"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "spacestatus"
}

func (m *Module) Help() string {
	return "USAGE: !spacestatus"
}

func (m *Module) Defaults() {
	m.Notify = true
	m.Interval = "0h15m"
}

func (m *Module) Load(client *irc.Client) error {
	if len(m.URL) <= 0 {
		return nil
	}

	duration, err := time.ParseDuration(m.Interval)
	if err != nil {
		return err
	}

	if err := m.updateHandler(client); err != nil {
		return err
	}

	if m.api.API != apiVersion {
		return errors.New("unsupported spaceapi version")
	}

	go func(c *irc.Client) {
		for {
			time.Sleep(duration)
			m.updateHandler(c)
		}
	}(client)

	client.CmdHook("privmsg", m.statusCmd)
	return nil
}

func (m *Module) updateHandler(client *irc.Client) error {
	var oldState bool
	if m.api == nil {
		oldState = false
	} else {
		oldState = m.api.State.Open
	}

	firstPoll := m.api == nil
	if err := m.pollStatus(); err != nil {
		return err
	}

	newState := m.api.State.Open
	if newState != oldState && m.Notify && !firstPoll {
		m.notify(client, newState)
	}

	return nil
}

func (m *Module) pollStatus() error {
	resp, err := http.Get(m.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &m.api); err != nil {
		return err
	}

	return nil
}

func (m *Module) statusCmd(client *irc.Client, msg irc.Message) error {
	if msg.Data != "!spacestatus" {
		return nil
	} else if m.api == nil {
		return client.Write("NOTICE %s :Status currently unknown.",
			msg.Receiver)
	}

	var state string
	if m.api.State.Open {
		state = "open"
	} else {
		state = "closed"
	}

	return client.Write("NOTICE %s :%s is currently %s",
		msg.Receiver, m.api.Space, state)
}

func (m *Module) notify(client *irc.Client, open bool) {
	var oldState, newState string
	if open {
		oldState = "closed"
		newState = "open"
	} else {
		oldState = "open"
		newState = "closed"
	}

	for _, ch := range client.Channels {
		client.Write("NOTICE %s :%s changed door status from %s to %s",
			ch, m.api.Space, oldState, newState)
	}
}

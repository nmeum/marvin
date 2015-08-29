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

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type config struct {
	// Nickname of the irc bot.
	Nick string `json:"nickname"`

	// Realname of the irc bot.
	Name string `json:"realname"`

	// Hostname of the irc server.
	Host string `json:"hostname"`

	// Port to connect to.
	Port int `json:"port"`

	// Path to directory containing module configs.
	Conf string `json:"configs"`

	// Path to SSL cert (if any).
	Cert string `json:"cert"`

	// List of channels to connect to.
	Chan []string `json:"channels"`
}

func confDefaults() Config {
	return Config{
		Nick: "marvin",
		Name: "marvin IRC Bot",
		Host: "chat.freenode.net",
		Port: 6667,
		Conf: filepath.Join(os.Getenv("HOME"), appName),
	}
}

func readConfig(path string) (c Config, err error) {
	c = confDefaults()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &c); err != nil {
		return
	}

	return
}

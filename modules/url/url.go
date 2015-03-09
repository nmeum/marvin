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

package url

import (
	"errors"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	urlRegex   = `(http|https)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~])*`
	extractErr = errors.New("couldn't extract title")
)

type Module struct {
	Regex   string   `json:"regex"`
	Exclude []string `json:"exclude"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "url"
}

func (m *Module) Help() string {
	return "Displays HTML titles for HTTP links."
}

func (m *Module) Defaults() {
	m.Regex = urlRegex
}

func (m *Module) Load(client *irc.Client) error {
	regex := regexp.MustCompile(m.Regex)
	client.CmdHook("privmsg", func(c *irc.Client, msg irc.Message) error {
		link := regex.FindString(msg.Data)
		if len(link) <= 0 {
			return nil
		}

		uri, err := url.Parse(link)
		if err != nil {
			return err
		}

		if err == nil && !m.isExcluded(uri.Host) {
			title, err := m.extractTitle(link)
			if err == nil {
				c.Write("NOTICE %s :Page title: %s", msg.Receiver, title)
			}
		}

		return nil
	})

	return nil
}

func (m *Module) isExcluded(host string) bool {
	for _, h := range m.Exclude {
		if host == h {
			return true
		}
	}

	return false
}

func (m *Module) extractTitle(uri string) (title string, err error) {
	resp, err := http.Get(uri)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	regex := regexp.MustCompile("(?is)<title>(.+)</title>")
	match := regex.FindSubmatch(body)

	if len(match) < 2 {
		err = extractErr
		return
	}

	title = m.normalize(title)
	if len(title) <= 0 {
		err = extractErr
		return
	}

	return
}

func (m *Module) normalize(title string) string {
	normalized := html.UnescapeString(strings.Replace(title, "\n", " ", -1))
	for strings.Contains(normalized, "  ") {
		normalized = strings.Replace(normalized, "  ", " ", -1)
	}

	return strings.TrimSpace(normalized)
}

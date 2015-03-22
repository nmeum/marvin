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
	"fmt"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"golang.org/x/net/html"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Module struct {
	regex    *regexp.Regexp
	RegexStr string `json:"regex"`
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
	m.RegexStr = `(?i)\b((http|https)\://(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))`
}

func (m *Module) Load(client *irc.Client) error {
	regex, err := regexp.Compile(m.RegexStr)
	if err != nil {
		return err
	}

	m.regex = regex
	client.CmdHook("privmsg", m.urlCmd)

	return nil
}

func (m *Module) urlCmd(client *irc.Client, msg irc.Message) error {
	url := m.regex.FindString(msg.Data)
	if len(url) <= 0 {
		return nil
	}

	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	resp.Body.Close() // HEAD response doesn't have a body

	info := m.infoString(resp)
	if len(info) <= 0 {
		return nil
	}

	return client.Write("NOTICE %s :%s", msg.Receiver, info)
}

func (m *Module) infoString(resp *http.Response) string {
	var mtype string
	var infos []string

	ctype := resp.Header.Get("Content-Type")
	if len(ctype) > 0 {
		m, _, err := mime.ParseMediaType(ctype)
		if err == nil {
			mtype = m
			infos = append(infos, fmt.Sprintf("Type: %s", mtype))
		}
	}

	csize := resp.Header.Get("Content-Length")
	if len(csize) > 0 {
		size, err := strconv.Atoi(csize)
		if err == nil {
			infos = append(infos, fmt.Sprintf("Size: %s", m.humanize(size)))
		}
	}

	if mtype == "text/html" {
		title, err := m.extractTitle(resp.Request.URL.String())
		if err == nil {
			infos = append(infos, fmt.Sprintf("Title: %s", title))
		}
	}

	info := strings.Join(infos, " | ")
	if len(info) > 0 {
		info = fmt.Sprintf("%s -- %s", strings.ToUpper(m.Name()), info)
	}

	return info
}

func (m *Module) extractTitle(url string) (title string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	var parseFunc func(n *html.Node)
	parseFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			child := n.FirstChild
			if child != nil {
				title = html.UnescapeString(child.Data)
			} else {
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseFunc(c)
		}
	}

	parseFunc(doc)
	title = m.sanitize(title)
	if len(title) <= 0 {
		err = errors.New("couldn't extract title")
		return
	}

	return
}

func (m *Module) sanitize(input string) string {
	mfunc := func(r rune) rune {
		if !unicode.IsPrint(r) {
			return ' '
		}

		return r
	}

	sanitized := strings.Map(mfunc, input)
	for strings.Contains(sanitized, "  ") {
		sanitized = strings.Replace(sanitized, "  ", " ", -1)
	}

	return strings.TrimSpace(sanitized)
}

func (m *Module) humanize(count int) string {
	switch {
	case count > (1 << 40):
		return fmt.Sprintf("%v TiB", count/(1<<40))
	case count > (1 << 30):
		return fmt.Sprintf("%v GiB", count/(1<<30))
	case count > (1 << 20):
		return fmt.Sprintf("%v MiB", count/(1<<20))
	case count > (1 << 10):
		return fmt.Sprintf("%v KiB", count/(1<<10))
	default:
		return fmt.Sprintf("%v B", count)
	}
}

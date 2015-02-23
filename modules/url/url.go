package url

import (
	"errors"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	urlRegex = `^(http|https)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~])*$`
)

type Module struct {
	// TODO
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "url"
}

func (m *Module) Help() string {
	return "Displays HTML titles for HTTP links"
}

func (m *Module) Load(client *irc.Client) error {
	regex := regexp.MustCompile(urlRegex)
	client.CmdHook("privmsg", func(c *irc.Client, msg irc.Message) error {
		if !regex.MatchString(msg.Data) {
			return nil
		}

		title, err := m.extractTitle(msg.Data)
		if err == nil {
			c.Write("NOTICE %s :Page title: %q", msg.Receiver, title)
		}

		return nil
	})

	return nil
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

	regex := regexp.MustCompile("<title>(.*)</title>")
	match := regex.FindStringSubmatch(string(body))

	if len(match) <= 0 {
		err = errors.New("Couldn't extract title")
	}

	title = html.UnescapeString(match[1])
	return
}

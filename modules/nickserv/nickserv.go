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
		if msg.Sender.Name != m.NickServ {
			return nil
		}

		if !strings.Contains(msg.Data, m.Keyword) {
			return nil

		}

		return c.Write("PRIVMSG %s :identify %s",
			m.NickServ, m.Password)
	})

	return nil
}

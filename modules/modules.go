package modules

import (
	"fmt"
	"github.com/nmeum/marvin/irc"
	"strings"
)

type Module interface {
	Name() string
	Help() string
	Load(*irc.Client) error
}

type ModuleSet struct {
	client  *irc.Client
	modules []Module
}

func NewModuleSet(client *irc.Client) *ModuleSet {
	return &ModuleSet{client: client}
}

func (m *ModuleSet) Register(module Module) {
	m.modules = append(m.modules, module)
}

func (m *ModuleSet) LoadAll() error {
	for _, module := range m.modules {
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
		strings.Join(names, " "))

	return client.Write("PRIVMSG %s :%s", msg.Receiver, help)
}

func (m *ModuleSet) moduleCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Split(msg.Data, " ")
	if len(splited) < 2 || splited[0] != "!help" {
		return nil
	}

	name := strings.ToLower(splited[1])
	module := m.findModule(name)
	if module == nil {
		return client.Write("PRIVMSG %s :Module %q isn't installed",
			msg.Receiver, name)
	}

	return client.Write("PRIVMSG %s :%s module: %s",
		msg.Receiver, module.Name(), module.Help())
}

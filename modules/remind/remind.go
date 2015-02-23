package remind

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"strings"
	"time"
)

type Module struct {
	TimeLimit int `json:"limit"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "remind"
}

func (m *Module) Help() string {
	return "!remind DURATION MSG"
}

func (m *Module) Defaults() {
	m.TimeLimit = 10
}

func (m *Module) Load(client *irc.Client) error {
	client.CmdHook("privmsg", func(c *irc.Client, msg irc.Message) error {
		splited := strings.Split(msg.Data, " ")
		if len(splited) < 3 || splited[0] != "!remind" {
			return nil
		}

		duration, err := time.ParseDuration(splited[1])
		if err != nil {
			return err
		}

		limit := (time.Duration)(m.TimeLimit) * time.Hour
		if duration > limit {
			return c.Write("NOTICE %s :%d exceeds the limit of %d",
				duration.Hours(), limit.Hours())
		}

		time.AfterFunc(duration, func() {
			c.Write("PRIVMSG %s :Reminder: %s",
				msg.Sender.Name, splited[2])
		})

		return c.Write("NOTICE %s :Reminder setup for %s",
			msg.Receiver, splited[1])
	})

	return nil
}

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

		rawDuration := splited[1]
		duration, err := time.ParseDuration(rawDuration)
		if err != nil {
			return err
		}

		limit := (time.Duration)(m.TimeLimit) * time.Hour
		if duration > limit {
			return c.Write("NOTICE %s :%d exceeds the limit of %d",
				duration.Hours(), limit.Hours())
		}

		reminder := strings.Join(splited[2:], " ")
		time.AfterFunc(duration, func() {
			c.Write("PRIVMSG %s :Reminder: %s",
				msg.Sender.Name, reminder)
		})

		return c.Write("NOTICE %s :Reminder setup for %s",
			msg.Receiver, rawDuration)
	})

	return nil
}

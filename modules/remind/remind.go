package remind

import (
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"strings"
	"time"
)

type Module struct {
	TimeLimit int `json:"time_limit"`
	UserLimit int `json:"user_limit"`
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
	m.UserLimit = 3
}

func (m *Module) Load(client *irc.Client) error {
	users := make(map[string]int)
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
			return c.Write("NOTICE %s :%v hours exceeds the limit of %v hours",
				msg.Receiver, duration.Hours(), limit.Hours())
		}

		if users[msg.Sender.Host] >= m.UserLimit {
			return c.Write("NOTICE %s :You can only run %d reminders at a time",
				msg.Receiver, m.UserLimit)
		}

		users[msg.Sender.Host]++
		reminder := strings.Join(splited[2:], " ")
		time.AfterFunc(duration, func() {
			users[msg.Sender.Host]--
			c.Write("PRIVMSG %s :Reminder: %s",
				msg.Sender.Name, reminder)
		})

		return c.Write("NOTICE %s :Reminder setup for %s",
			msg.Receiver, rawDuration)
	})

	return nil
}

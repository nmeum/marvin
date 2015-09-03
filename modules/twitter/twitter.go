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

package twitter

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"net/url"
	"strconv"
	"strings"
)

type Module struct {
	api               *anaconda.TwitterApi
	ReadOnly          bool   `json:"read_only"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

func Init(moduleSet *modules.ModuleSet) {
	moduleSet.Register(new(Module))
}

func (m *Module) Name() string {
	return "twitter"
}

func (m *Module) Help() string {
	return "USAGE: !tweet TEXT || !reply ID TEXT || !retweet ID || !favorite ID"
}

func (m *Module) Defaults() {
	m.ReadOnly = false
}

func (m *Module) Load(client *irc.Client) error {
	anaconda.SetConsumerKey(m.ConsumerKey)
	anaconda.SetConsumerSecret(m.ConsumerSecret)
	m.api = anaconda.NewTwitterApi(m.AccessToken, m.AccessTokenSecret)

	if !m.ReadOnly {
		client.CmdHook("privmsg", m.tweetCmd)
		client.CmdHook("privmsg", m.replyCmd)
		client.CmdHook("privmsg", m.retweetCmd)
		client.CmdHook("privmsg", m.favoriteCmd)
	}

	values := url.Values{}
	values.Add("replies", "all")
	values.Add("with", "user")

	go func(client *irc.Client, values url.Values) {
		for {
			m.streamHandler(client, values)
		}
	}(client, values)

	return nil
}

func (m *Module) tweetCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!tweet" || !client.Connected(msg.Receiver) {
		return nil
	}

	status := strings.Join(splited[1:], " ")
	if _, err := m.api.PostTweet(status, url.Values{}); err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return nil
}

func (m *Module) replyCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 3 || splited[0] != "!reply" || !client.Connected(msg.Receiver) {
		return nil
	}

	values := url.Values{}
	values.Add("in_reply_to_status_id", splited[1])

	status := strings.Join(splited[2:], " ")
	if !strings.Contains(status, "@") {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, "A reply must contain a @mention")
	}

	if _, err := m.api.PostTweet(status, values); err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return nil
}

func (m *Module) retweetCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!retweet" || !client.Connected(msg.Receiver) {
		return nil
	}

	id, err := strconv.Atoi(splited[1])
	if err != nil {
		return err
	}

	if _, err := m.api.Retweet(int64(id), false); err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return nil
}

func (m *Module) favoriteCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!favorite" || !client.Connected(msg.Receiver) {
		return nil
	}

	id, err := strconv.Atoi(splited[1])
	if err != nil {
		return err
	}

	if _, err := m.api.Favorite(int64(id)); err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return nil
}

func (m *Module) streamHandler(client *irc.Client, values url.Values) {
	stream := m.api.UserStream(values)
	for {
		select {
		case event := <-stream.C:
			if t := m.formatEvent(event); len(t) > 0 {
				m.notify(client, t)
			}
		case <-stream.Quit:
			break
		}
	}
}

func (m *Module) formatEvent(event interface{}) string {
	var msg string
	switch t := event.(type) {
	case anaconda.ApiError:
		msg = fmt.Sprintf("Twitter API error %d: %s", t.StatusCode, t.Decoded.Error())
	case anaconda.Tweet:
		msg = fmt.Sprintf("Tweet %d by %s: %s", t.Id, t.User.ScreenName, t.Text)
	case anaconda.StatusDeletionNotice:
		msg = fmt.Sprintf("Tweet %d has been deleted", t.Id)
	case anaconda.EventTweet:
		if t.Event.Event != "favorite" {
			break
		}

		msg = fmt.Sprintf("%s favorited tweet %d by %s: %s",
			t.Source.ScreenName, t.TargetObject.Id, t.Target.ScreenName, t.TargetObject.Text)
	}

	return msg
}

func (m *Module) notify(client *irc.Client, text string) {
	for _, ch := range client.Channels {
		client.Write("NOTICE %s :%s", ch, text)
	}
}

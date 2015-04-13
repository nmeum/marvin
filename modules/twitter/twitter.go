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
	"github.com/ChimeraCoder/anaconda"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"html"
	"net/url"
	"strconv"
	"strings"
)

type Module struct {
	api               *anaconda.TwitterApi
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
	return
}

func (m *Module) Load(client *irc.Client) error {
	anaconda.SetConsumerKey(m.ConsumerKey)
	anaconda.SetConsumerSecret(m.ConsumerSecret)
	m.api = anaconda.NewTwitterApi(m.AccessToken, m.AccessTokenSecret)

	client.CmdHook("privmsg", m.tweetCmd)
	client.CmdHook("privmsg", m.replyCmd)
	client.CmdHook("privmsg", m.retweetCmd)
	client.CmdHook("privmsg", m.favoriteCmd)

	values := url.Values{}
	values.Add("replies", "all")
	values.Add("with", "user")

	stream := m.api.UserStream(values)
	go func(c *irc.Client, s anaconda.Stream) {
		for i := range s.C {
			t, ok := i.(anaconda.Tweet)
			if ok {
				m.notify(c, t)
			}
		}
	}(client, stream)

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

	tweet, err := m.api.Favorite(int64(id))
	if err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return client.Write("NOTICE %s :Favorited tweet %d by %s: %s",
		msg.Receiver, tweet.Id, tweet.User.ScreenName, m.sanitize(tweet.Text))
}

func (m *Module) notify(client *irc.Client, tweet anaconda.Tweet) {
	for _, ch := range client.Channels {
		client.Write("NOTICE %s :Tweet %d by %s: %s",
			ch, tweet.Id, tweet.User.ScreenName, m.sanitize(tweet.Text))
	}
}

func (m *Module) sanitize(text string) string {
	sanitized := html.UnescapeString(text)
	sanitized = strings.Replace(sanitized, "\n", " ", -1)
	return strings.TrimSpace(sanitized)
}

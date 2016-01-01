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
	"html"
	"net/url"
	"strconv"
	"strings"
)

// Maximum amount of characters allowed in a tweet.
const maxChars = 140

type Module struct {
	api               *anaconda.TwitterApi
	user              anaconda.User
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
	return "USAGE: !tweet TEXT || !reply ID TEXT || !directmsg USER TEXT || !retweet ID || !favorite ID || !stat ID"
}

func (m *Module) Defaults() {
	m.ReadOnly = false
}

func (m *Module) Load(client *irc.Client) error {
	anaconda.SetConsumerKey(m.ConsumerKey)
	anaconda.SetConsumerSecret(m.ConsumerSecret)

	m.api = anaconda.NewTwitterApi(m.AccessToken, m.AccessTokenSecret)
	client.CmdHook("privmsg", m.statCmd)

	if !m.ReadOnly {
		client.CmdHook("privmsg", m.tweetCmd)
		client.CmdHook("privmsg", m.replyCmd)
		client.CmdHook("privmsg", m.retweetCmd)
		client.CmdHook("privmsg", m.favoriteCmd)
		client.CmdHook("privmsg", m.directMsgCmd)
	}

	values := url.Values{}
	values.Add("skip_status", "true")

	user, err := m.api.GetSelf(values)
	if err != nil {
		return err
	} else {
		m.user = user
	}

	values = url.Values{}
	values.Add("replies", "all")
	values.Add("with", "user")

	go func(c *irc.Client, v url.Values) {
		for {
			m.streamHandler(c, v)
		}
	}(client, values)

	return nil
}

func (m *Module) tweet(t string, v url.Values, c *irc.Client, p irc.Message) error {
	_, err := m.api.PostTweet(t, v)
	if err != nil && len(t) > maxChars {
		return c.Write("NOTICE %s :ERROR: Tweet is too long, remove %d characters",
			p.Receiver, len(t)-maxChars)
	} else if err != nil {
		return c.Write("NOTICE %s :ERROR: %s", p.Receiver, err.Error())
	} else {
		return nil
	}
}

func (m *Module) tweetCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!tweet" || !client.Connected(msg.Receiver) {
		return nil
	}

	status := strings.Join(splited[1:], " ")
	return m.tweet(status, url.Values{}, client, msg)
}

func (m *Module) replyCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 3 || splited[0] != "!reply" || !client.Connected(msg.Receiver) {
		return nil
	}

	status := strings.Join(splited[2:], " ")
	if !strings.Contains(status, "@") {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, "A reply must contain an @mention")
	}

	values := url.Values{}
	values.Add("in_reply_to_status_id", splited[1])

	return m.tweet(status, values, client, msg)
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

func (m *Module) directMsgCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 3 || splited[0] != "!directmsg" || !client.Connected(msg.Receiver) {
		return nil
	}

	scname := splited[1]
	status := strings.Join(splited[2:], " ")

	if _, err := m.api.PostDMToScreenName(status, scname); err != nil {
		return client.Write("NOTICE %s :ERROR: %s",
			msg.Receiver, err.Error())
	}

	return nil
}

func (m *Module) statCmd(client *irc.Client, msg irc.Message) error {
	splited := strings.Fields(msg.Data)
	if len(splited) < 2 || splited[0] != "!stat" || !client.Connected(msg.Receiver) {
		return nil
	}

	id, err := strconv.Atoi(splited[1])
	if err != nil {
		return err
	}

	tweet, err := m.api.GetTweet(int64(id), url.Values{})
	if err != nil {
		return err
	}

	return client.Write("NOTICE %s :Stats for tweet %d by %s: ↻ %d ★ %d",
		msg.Receiver, tweet.Id, tweet.User.ScreenName, tweet.RetweetCount, tweet.FavoriteCount)
}

func (m *Module) streamHandler(client *irc.Client, values url.Values) {
	stream := m.api.UserStream(values)
	for {
		select {
		case event, ok := <-stream.C:
			if !ok {
				break
			}

			if t := m.formatEvent(event); len(t) > 0 {
				m.notify(client, t)
			}
		}
	}

	stream.Stop()
}

func (m *Module) formatEvent(event interface{}) string {
	var msg string
	switch t := event.(type) {
	case anaconda.ApiError:
		msg = fmt.Sprintf("Twitter API error %d: %s", t.StatusCode, t.Decoded.Error())
	case anaconda.StatusDeletionNotice:
		msg = fmt.Sprintf("Tweet %d has been deleted", t.Id)
	case anaconda.DirectMessage:
		msg = fmt.Sprintf("Direct message %d by %s send to %s: %s", t.Id,
			t.SenderScreenName, t.RecipientScreenName, html.UnescapeString(t.Text))
	case anaconda.Tweet:
		if t.RetweetedStatus != nil && t.User.Id != m.user.Id {
			break
		}

		msg = fmt.Sprintf("Tweet %d by %s: %s", t.Id, t.User.ScreenName,
			html.UnescapeString(t.Text))
	case anaconda.EventTweet:
		if t.Event.Event != "favorite" || t.Source.Id != m.user.Id {
			break
		}

		text := html.UnescapeString(t.TargetObject.Text)
		msg = fmt.Sprintf("%s favorited tweet %d by %s: %s",
			t.Source.ScreenName, t.TargetObject.Id, t.Target.ScreenName, text)
	}

	return msg
}

func (m *Module) notify(client *irc.Client, text string) {
	for _, ch := range client.Channels {
		client.Write("NOTICE %s :%s", ch, text)
	}
}

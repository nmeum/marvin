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
	"net/url"
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
	return "Enables interaction with the socialmedia website twitter."
}

func (m *Module) Defaults() {
	return
}

func (m *Module) Load(client *irc.Client) error {
	anaconda.SetConsumerKey(m.ConsumerKey)
	anaconda.SetConsumerSecret(m.ConsumerSecret)
	m.api = anaconda.NewTwitterApi(m.AccessToken, m.AccessTokenSecret)

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

func (m *Module) notify(client *irc.Client, tweet anaconda.Tweet) {
	for _, ch := range client.Channels {
		client.Write("NOTICE %s :Tweet %d: %s",
			ch, tweet.Id, tweet.Text)
	}
}

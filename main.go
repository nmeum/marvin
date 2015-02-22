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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/nmeum/marvin/irc"
	"log"
	"net"
	"os"
	"time"
)

const (
	appName  = "marvin"
	appUsage = "[options] CHANNEL..."
)

var (
	nick = flag.String("n", "marvin", "nickname")
	name = flag.String("r", "Marvin Bot", "realname")
	host = flag.String("h", "irc.hackint.eu", "host")
	port = flag.Int("p", 6667, "port")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stderr, fmt.Sprintf("%s:", appName), 0)
	if flag.NArg() < 1 {
		logger.Fatalf("USAGE: %s %s", appName, appUsage)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		logger.Fatal(err)
	}

	defer conn.Close()

	for {
		ircBot := newBot(conn, flag.Args())
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			ircBot.Handle(line)
		}

		if err := scanner.Err(); err != nil {
			logger.Println(err)
		}

		var err error
		for i := 1; err != nil; i++ {
			conn, err = reconnect(conn)
			if err != nil {
				logger.Println(err)
				time.Sleep((time.Duration)(i*3) * time.Second)
			}
			defer conn.Close()
		}
	}
}

func newBot(conn net.Conn, channels []string) *irc.Client {
	client := irc.NewClient(conn)
	client.CmdHook("ping", func(c *irc.Client, m *irc.Message) error {
		return c.Write("PONG %s", m.Data)
	})

	client.CmdHook("001", func(c *irc.Client, m *irc.Message) error {
		for _, ch := range channels {
			if err := c.Write("JOIN %s", ch); err != nil {
				return err
			}
		}

		return nil
	})

	client.Write("USER %s localhost * :%s", *nick, *name)
	client.Write("NICK %s", *nick)

	return client
}

func reconnect(c net.Conn) (conn net.Conn, err error) {
	addr := c.RemoteAddr()
	c.Close()

	conn, err = net.Dial("tcp", addr.String())
	if err != nil {
		return
	}

	return
}

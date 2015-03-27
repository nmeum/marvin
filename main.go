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
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/nmeum/marvin/irc"
	"github.com/nmeum/marvin/modules"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	cert = flag.String("c", "", "certificate")
	verb = flag.Bool("v", false, "verbose output")
	port = flag.Int("p", 6697, "port")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stderr, "ERROR: ", 0)
	if flag.NArg() < 1 {
		logger.Fatalf("USAGE: %s %s", appName, appUsage)
	}

	conn, err := connect("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	errChan := make(chan error)
	go func() {
		for err := range errChan {
			logger.Println(err)
		}
	}()

	for {
		ircBot, err := setup(&conn, flag.Args())
		if err != nil {
			logger.Println(err)
		}

		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				logger.Println(err)
				break
			}

			line = strings.Trim(line, "\n")
			line = strings.Trim(line, "\r")

			if *verb {
				fmt.Println(line)
			}

			ircBot.Handle(line, errChan)
		}

		conn = reconnect(conn)
	}
}

func setup(conn *net.Conn, channels []string) (client *irc.Client, err error) {
	client = irc.NewClient(conn)
	client.CmdHook("ping", func(c *irc.Client, m irc.Message) error {
		return c.Write("PONG %s", m.Data)
	})

	joinCmd := func(c *irc.Client, m irc.Message) error {
		time.Sleep(3 * time.Second)
		return c.Write("JOIN %s", strings.Join(channels, ","))
	}

	client.CmdHook("001", joinCmd)
	client.CmdHook("kick", joinCmd)

	client.Write("USER %s %s * :%s", *nick, *host, *name)
	client.Write("NICK %s", *nick)

	if err = initializeModules(client); err != nil {
		return
	}

	return
}

func initializeModules(c *irc.Client) error {
	config := os.Getenv("XDG_CONFIG_HOME")
	if len(config) <= 0 {
		user, err := user.Current()
		if err != nil {
			return err
		}

		config = filepath.Join(user.HomeDir, ".config")
	}

	moduleSet := modules.NewModuleSet(c, filepath.Join(config, appName))
	for _, fn := range moduleInits {
		fn(moduleSet)
	}

	return moduleSet.LoadAll()
}

func connect(network, address string) (conn net.Conn, err error) {
	if len(*cert) >= 1 {
		certFile, err := ioutil.ReadFile(*cert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(certFile)

		config := &tls.Config{RootCAs: caCertPool}
		return tls.Dial(network, address, config)
	}

	return net.Dial(network, address)
}

func reconnect(c net.Conn) (conn net.Conn) {
	addr := c.RemoteAddr()
	c.Close()

	var err error
	for i := 1; ; i++ {
		conn, err = connect(addr.Network(), addr.String())
		if err == nil {
			break
		}

		time.Sleep(time.Duration(i*3) * time.Second)
	}

	return
}

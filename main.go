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
	"strings"
	"time"
)

const (
	appName = "marvin"
)

var (
	conf = flag.String("c", "marvin.json", "configuration file")
	verb = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stderr, "ERROR: ", 0)

	config, err := readConfig(*conf)
	if err != nil && !os.IsNotExist(err) {
		logger.Fatal(err)
	}

	conn, err := connect(config)
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

	ircBot, err := setup(conn, config)
	if err != nil {
		logger.Fatal(err)
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
}

func setup(conn net.Conn, config Config) (client *irc.Client, err error) {
	client = irc.NewClient(conn)
	client.CmdHook("001", func(c *irc.Client, m irc.Message) error {
		time.Sleep(3 * time.Second) // Wait for NickServ etc
		return c.Write("JOIN %s", strings.Join(config.Chan, ","))
	})

	client.CmdHook("kick", func(c *irc.Client, m irc.Message) error {
		params := strings.Fields(m.Receiver)
		return c.Write("JOIN %s", params[0])
	})

	client.Write("USER %s %s * :%s", config.Nick, config.Host, config.Name)
	client.Write("NICK %s", config.Nick)

	moduleSet := modules.NewModuleSet(client, config.Conf)
	for _, fn := range moduleInits {
		fn(moduleSet)
	}

	return client, moduleSet.LoadAll()
}

func connect(config Config) (conn net.Conn, err error) {
	netw := "tcp"
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	if len(config.Cert) >= 1 {
		certFile, err := ioutil.ReadFile(config.Cert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(certFile)

		config := &tls.Config{RootCAs: caCertPool}
		return tls.Dial(netw, addr, config)
	}

	return net.Dial(netw, addr)
}

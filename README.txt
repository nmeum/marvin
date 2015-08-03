marvin - A modular irc bot written in go.

INSTALLATION
	To install, run `go get -u github.com/nmeum/marvin`.

USAGE
	marvin accepts three command line flags: '-h', '-v', and '-c'.

	When '-h' is used marvin writes the help message to stderr and
	exits with exit status 2. With '-v' marvin writes everything it
	reads from the TCP socket to stdout. The last flag '-c' allows
	the caller to specify the path of a configuration file described
	in greater detail below.

CONFIGURATION
	marvin is configured using a small json file. There is a core
	configuration file which can be specified with the '-c' command
	line flag and in addition to that there is a json configuration
	file for each standalone module.

	The core configuration file allows you to specify mandatory
	information for the bot, e.g. which network to connect to, which
	username to use, which channels to join, et cetera. The
	available configuration variables are documented in the `Config
	struct` defined in the file `config.go`.

MODULES
	marvin is a very modular irc bot. Each module has its own
	configuration file and can be enabled or disabled. Most modules
	are enabled by default.

	To enable a module you have to add its module initialization
	function to the `moduleInits` slice defined in the file
	`modules.go`. The module initialization function is mostly
	called `Init`. To add it to the array you need to import it
	first of cause.

	Modules are configured in the specified `configs` directory
	which defaults to `$HOME/marvin`. You can specify a different
	configuration directory in the core configuration file. The
	available configuration variables are defined in the individual
	module, please consult the code to get an overview of the
	available options.

LICENSE
	This program is free software: you can redistribute it and/or
	modify it under the terms of the GNU Affero General Public
	License as published by the Free Software Foundation, either
	version 3 of the License, or (at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
	Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public
	License along with this program. If not, see
	<http://www.gnu.org/licenses/>.

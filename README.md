# marvin - A modular irc bot written in go.

### Installation
To install, run `go get -u github.com/nmeum/marvin`.

### Usage
marvin accepts three command line flags: '-h', '-v', and '-c'.

When '-h' is used marvin writes the help message to stderr and exits with exit status 2. With '-v' marvin writes everything it reads from the TCP socket to stdout.
The last flag '-c' allows the caller to specify the path of a configuration file described in greater detail below.

### Configuration
marvin is configured using a small json file. There is a core configuration file which can be specified with the '-c' command line flag and in addition to that there is a json configuration file for each standalone module.
The core configuration file allows you to specify mandatory information for the bot, e.g. which network to connect to, which username to use, which channels to join, et cetera.
The available configuration variables are documented in the `config struct` defined in the file `config.go`.

### Modules
marvin is a very modular irc bot. Each module has its own configuration file and can be enabled or disabled. Most modules are enabled by default.

To enable a module you have to add its module initialization function to the `moduleInits` slice defined in the file `modules.go`.
The module initialization function is mostly called `Init`. To add it to the array you need to import it first of cause.
Modules are configured in the specified `configs` directory which defaults to `$HOME/marvin`. You can specify a different configuration directory in the core configuration file.

The available configuration variables are defined in the individual module, please consult the code to get an overview of the available options.

### License
This program is release under the GNU AFFERO GENERAL PUBLIC LICENSE. You can check it [here](LICENSE)
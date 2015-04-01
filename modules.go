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
	"github.com/nmeum/marvin/modules"
	"github.com/nmeum/marvin/modules/feed"
	"github.com/nmeum/marvin/modules/nickserv"
	"github.com/nmeum/marvin/modules/remind"
	"github.com/nmeum/marvin/modules/time"
	"github.com/nmeum/marvin/modules/twitter"
	"github.com/nmeum/marvin/modules/url"
)

type moduleInit func(*modules.ModuleSet)

var moduleInits = []moduleInit{
	nickserv.Init,
	twitter.Init,
	remind.Init,
	time.Init,
	feed.Init,
	url.Init,
}

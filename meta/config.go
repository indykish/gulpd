/*
** Copyright [2013-2016] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */

package meta

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/megamsys/libgo/cmd"
)

const (
	// DefaultRiak is the default riak if one is not provided.
	DefaultRiak = "localhost:8087"

	// DefaultNSQ is the default nsqd if its not provided.
	DefaultNSQd = "localhost:4161"

	//DefaultDockerPath is the detault docker path
	DefaultDockerPath = "/var/lib/docker/containers/"
)

var MC *Config

// Config represents the meta configuration.
type Config struct {
	Home       string   `toml:"home"` //figured out from MEGAM_HOME variable
	Dir        string   `toml:"dir"`
	User       string   `toml:"user"`
	Riak       []string `toml:"riak"`
	NSQd       []string `toml:"nsqd"`
	DockerPath string   `toml:"docker_path"`
}

func (c Config) String() string {
	w := new(tabwriter.Writer)
	var b bytes.Buffer
	w.Init(&b, 0, 8, 0, '\t', 0)
	b.Write([]byte(cmd.Colorfy("Config:", "white", "", "bold") + "\t" +
		cmd.Colorfy("Meta", "green", "", "") + "\n"))
	b.Write([]byte("Home" + "\t" + c.Home + "\n"))
	b.Write([]byte("Dir" + "\t" + c.Dir + "\n"))
	b.Write([]byte("User" + "\t" + c.User + "\n"))
	b.Write([]byte("Riak" + "\t" + strings.Join(c.Riak, ",") + "\n"))
	b.Write([]byte("NSQd      " + "\t" + strings.Join(c.NSQd, ",") + "\n"))
	b.Write([]byte("DockerPath" + "\t" + c.DockerPath + "\n"))
	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func NewConfig() *Config {
	var homeDir string
	if os.Getenv("MEGAM_HOME") != "" {
		homeDir = os.Getenv("MEGAM_HOME")
	} else if u, err := user.Current(); err == nil {
		homeDir = u.HomeDir
	} else {
		return nil
	}

	defaultDir := filepath.Join(homeDir, "gulp")

	_ = os.MkdirAll(defaultDir, 0755)

	// Config represents the configuration format for the gulpd.
	return &Config{
		Home:       homeDir,
		Dir:        defaultDir,
		Riak:       []string{DefaultRiak},
		NSQd:       []string{DefaultNSQd},
		DockerPath: DefaultDockerPath,
	}
}

func (c *Config) MkGlobal() {
	MC = c
}

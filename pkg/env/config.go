// Joseph Bursey <jbursey@tevora.com>

package env

import(
	"fafo/pkg/httpclient"
)

const (
	DefaultFile string = "profiles/default.cfg"
)

type Config struct {
	NumWorkers    uint
	ClientCfg     httpclient.HttpCfg

	FuzzRecursive bool

	Seclists      string
	FuzzDirList   string
	FuzzFileList  string
}

func DefaultConfig() *Config {
	return &Config{
		NumWorkers:    8,
		ClientCfg:     *httpclient.DefaultConfig(),
		FuzzRecursive: true,
		FuzzDirList:   "Discovery/Web-Content/raft-medium-directories-lowercase.txt",
		FuzzFileList:  "Discovery/Web-Content/raft-medium-files.txt",
	}
}

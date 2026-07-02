// Joseph Bursey <jbursey@tevora.com>

package env

// TODO: Take config for individual seclists, Seclists location, etc.

type Config struct {
	NumWorkers    uint
	NumHttpCalls  int

	FuzzRecursive bool

	Seclists      string
	FuzzDirList   string
	FuzzFileList  string
}

func DefaultConfig() *Config {
	return &Config{
		NumWorkers:    8,
		NumHttpCalls:  5,
		FuzzRecursive: true,
		FuzzDirList:   "Discovery/Web-Content/raft-medium-directories-lowercase.txt",
		FuzzFileList:  "Discovery/Web-Content/raft-medium-files.txt",
	}
}

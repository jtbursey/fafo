// Joseph Bursey <jbursey@tevora.com>

package env

import(
    //"encoding/json"

    "fafo/pkg/httpclient"
)

const (
    DefaultConfigFile string = "profiles/default.cfg"
)

type Config struct {
    NumWorkers        uint
    ClientCfg         httpclient.HttpCfg

    FuzzRecursive     bool
    DisableScreenShot bool

    FindingsDir       string
    ScrShDir          string
    ScrShExt          string

    Seclists          string
}

func DefaultConfig() *Config {
    return &Config{
        NumWorkers:        8,
        ClientCfg:         *httpclient.DefaultConfig(),
        FuzzRecursive:     true,
        DisableScreenShot: false,
        ScrShExt:          "jpeg",
    }
}

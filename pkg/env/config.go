// Joseph Bursey <jbursey@tevora.com>

package env

import(
    "encoding/json"
    "os"

    "fafo/pkg/httpclient"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

const (
    DefaultConfigFile   string = "profiles/default.cfg"
    DefaultPayloadsFile string = "workflow/payloads.json"
    DefaultFindingsDir  string = "findings"
)

type Config struct {
    NumWorkers        uint               `json:"NumWorkers"`
    ClientCfg         httpclient.HttpCfg `json:"Client"`

    FuzzRecursive     bool               `json:"FuzzRecursive"`
    DisableScreenShot bool

    FindingsDir       string
    ScrShDir          string
    ScrShExt          string

    PayloadSrc        string             `json:"Payloads"`
    PayloadFiles      map[string]string
}

func DefaultConfig() *Config {
    return &Config{
        NumWorkers:        8,
        ClientCfg:         *httpclient.DefaultConfig(),
        FuzzRecursive:     true,
        DisableScreenShot: false,
        ScrShExt:          "jpeg",
        PayloadSrc:        DefaultPayloadsFile,
    }
}

func (c *Config) ParsePayloads() error {
    data, err := os.ReadFile(c.PayloadSrc)
    if err != nil {
        log.Errf("Failed to read config file %v: %v\n", c.PayloadSrc, err)
        return err
    }

    err = json.Unmarshal(data, &c.PayloadFiles)
    if err != nil {
        log.Errf("Failed to parse config json: %v\n", err)
        return err
    }

    return nil
}

func (c *Config) Parse(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        log.Errf("Failed to read config file %v: %v\n", filename, err)
        return err
    }
    
    err = json.Unmarshal(data, c)
    if err != nil {
        log.Errf("Failed to parse config json: %v\n", err)
        return err
    }

    err = c.ParsePayloads()
    if err != nil {
        return err
    }

    c.ClientCfg.PostParse()

    return nil
}

func (c *Config) Debug() {
    log.Logf(0, "%v\n", pretty.Config("Output", c.FindingsDir))
    log.Logf(0, "%v\n", pretty.Config("Workers", c.NumWorkers))
    log.Logf(0, "%v\n", pretty.Config("FuzzRecursive", c.FuzzRecursive))

    c.ClientCfg.Debug()

    log.Logf(2, "\n%v\n", pretty.Config("Payloads", c.PayloadSrc))
    for key, value := range c.PayloadFiles {
        log.Logf(2, "%v\n", pretty.Config("  "+key, value))
    }
}

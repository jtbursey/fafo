// Joseph Bursey <jbursey@tevora.com>

package config

import (
    "encoding/json"
    "fmt"
    "os"

    "fafo/pkg/fs"
    "fafo/pkg/httpclient"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

const (
    DefaultConfigFile   string = "profiles/default.cfg"
    DefaultPayloadsFile string = "workflow/payloads.json"
    DefaultFindingsDir  string = "findings"
)

var (
    DefaulSeclistFiles map[string]string = map[string]string{
        "DirectoryFuzzList": "Discovery/Web-Content/raft-medium-directories-lowercase.txt",
        "FileFuzzList":      "Discovery/Web-Content/raft-medium-files.txt",
    }
)

type Config struct {
    SelfFile          string
    NumWorkers        uint               `json:"NumWorkers"`
    ClientCfg         httpclient.HttpCfg `json:"Client"`

    FuzzRecursive     bool               `json:"FuzzRecursive"`
    DisableScreenShot bool

    FindingsDir       string
    ScrShDir          string
    ScrShExt          string

    Seclists          string             `json:"Seclists"`
    PayloadSrc        string             `json:"Payloads"`
    PayloadFiles      map[string]string
}

func DefaultConfig() *Config {
    return &Config{
        NumWorkers:        4,
        ClientCfg:         *httpclient.DefaultConfig(),
        FuzzRecursive:     true,
        DisableScreenShot: false,
        ScrShExt:          "jpeg",
        PayloadSrc:        DefaultPayloadsFile,
    }
}

func (c *Config) ParsePayloads() error {
    if !fs.Exists(c.PayloadSrc) {
        return nil
    }

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

func (c *Config) GetAsFilename(fn string) (string, error) {
    var ok bool
    var filename string
    if !fs.Exists(fn) {
        filename, ok = c.PayloadFiles[fn]
        if !ok {
            return "", fmt.Errorf("Invalid payload file key / filename: %v", fn)
        }
    }
    return filename, nil
}

func (c *Config) NeedSeclists() string {
    if len(c.Seclists) <= 0 || !fs.Exists(c.Seclists) {
        log.Logf(0, "Config \"Seclists\" is not set yet. It can be optionally set in %v.\n", c.SelfFile)
        c.Seclists = fs.GetFileFromStdio("Path to SecLists")
    }
    return c.Seclists
}

func (c *Config) Parse() error {
    data, err := os.ReadFile(c.SelfFile)
    if err != nil {
        log.Errf("Failed to read config file %v: %v\n", c.SelfFile, err)
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

func (c *Config) WritePayloadsJSON() error {
    data, err := json.MarshalIndent(c.PayloadFiles, "", "    ")
    if err != nil {
        log.Errf("Failed to marshal payload files json: %v", err)
        return err
    }

    err = os.WriteFile(c.PayloadSrc, data, 0664)
    if err != nil {
        log.Errf("Failed to write json to %v: %v", err)
        return err
    }

    return nil
}

func (c *Config) Debug() {
    if c.FindingsDir != "" {
        log.Logf(0, "%v\n", pretty.Config("Output", c.FindingsDir))
    }
    log.Logf(0, "%v\n", pretty.Config("Workers", c.NumWorkers))
    log.Logf(0, "%v\n", pretty.Config("FuzzRecursive", c.FuzzRecursive))

    c.ClientCfg.Debug()

    log.Logf(2, "\n%v\n", pretty.Config("Payloads", c.PayloadSrc))
    for key, value := range c.PayloadFiles {
        log.Logf(2, "%v\n", pretty.Config("  "+key, value))
    }
}

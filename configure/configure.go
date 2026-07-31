// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"

    "fafo/pkg/config"
    "fafo/pkg/env"
    "fafo/pkg/log"
)

var (
    flagC = flag.String("c", "", "The `Config File` to use")
    flagConfig = flag.String("config", config.DefaultConfigFile, "The `Config File` to use")
)

var (
    V0 int = log.V0
)

func main() {
    log.SetVerb(config.VPayloads)
    flag.Parse()

    log.Greeting("Configuring...")

    cfg := config.DefaultConfig()
    if *flagC != "" {
        cfg.SelfFile = *flagC
    } else {
        cfg.SelfFile = *flagConfig
    }

    cfg.FindingsDir = config.DefaultFindingsDir

    if err := cfg.Parse(); err != nil {
        log.Errf("Failed to parse %v: %v\n", cfg.SelfFile, err)
        return
    }

    env := &env.Env{
        Cfg: *cfg,
    }

    if err := env.ParseActions(); err != nil {
        return
    }
    
    if err := env.Validate(); err != nil{
        log.Errf("Environment failed validation: %v\n", err)
        return
    }

    log.Log(V0, "\n")
    env.Debug()

    if err := env.Cfg.WritePayloadsJSON(); err != nil {
        log.Err("Configuration Failed.\n")
        return
    }
    log.Logf(V0, "Payloads written to %v\n", env.Cfg.PayloadSrc)
    log.Log(V0, "Configured.\n")
}
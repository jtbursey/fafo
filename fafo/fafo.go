// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"
    "fmt"
    "time"

    "fafo/pkg/env"
    "fafo/pkg/fact"
    "fafo/pkg/httpclient"
    "fafo/pkg/job"
    "fafo/pkg/log"
    "fafo/pkg/worker"
)

// Define the command line args here
var (
    flagURL = flag.String("url", "", "The base `URL` (domain) to hit")
    flagEP = flag.String("ep", "", "The specific `Endpoint` to hit (overrides URL)")
    flagPort = flag.Int("p", 443, "The `Port` on which to scan")
)

func Greeting() {
// ______                ___                            _    ______ _           _   _____       _
// |  ___|              / _ \                          | |   |  ___(_)         | | |  _  |     | |
// | |_ _   _ ________ / /_\ \_ __ ___  _   _ _ __   __| |   | |_   _ _ __   __| | | | | |_   _| |_
// |  _| | | |_  /_  / |  _  | '__/ _ \| | | | '_ \ / _` |   |  _| | | '_ \ / _` | | | | | | | | __|
// | | | |_| |/ / / /  | | | | | | (_) | |_| | | | | (_| |_  | |   | | | | | (_| | \ \_/ / |_| | |_
// \_|  \__,_/___/___| \_| |_/_|  \___/ \__,_|_| |_|\__,_( ) \_|   |_|_| |_|\__,_|  \___/ \__,_|\__|
//                                                       |/

    log.Log(0, "\n\n=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n")
    log.Log(0, "  ______                ___                            _    ______ _           _   _____       _\n")
    log.Log(0, "  |  ___|              / _ \\                          | |   |  ___(_)         | | |  _  |     | |\n")
    log.Log(0, "  | |_ _   _ ________ / /_\\ \\_ __ ___  _   _ _ __   __| |   | |_   _ _ __   __| | | | | |_   _| |_\n")
    log.Log(0, "  |  _| | | |_  /_  / |  _  | '__/ _ \\| | | | '_ \\ / _` |   |  _| | | '_ \\ / _` | | | | | | | | __|\n")
    log.Log(0, "  | | | |_| |/ / / /  | | | | | | (_) | |_| | | | | (_| |_  | |   | | | | | (_| | \\ \\_/ / |_| | |_\n")
    log.Log(0, "  \\_|  \\__,_/___/___| \\_| |_/_|  \\___/ \\__,_|_| |_|\\__,_( ) \\_|   |_|_| |_|\\__,_|  \\___/ \\__,_|\\__|\n")
    log.Logf(0, "   %-53v|/\n", "")
    log.Log(0, "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n\n\n")
}

func Loop(env *env.Env) {
    prefix := ""
    if log.Verb(3) {
        prefix = fmt.Sprintf("%-13v", "[Manager]: ")
    }
    for ;; {
        select {
        case t := <- env.FactCh:
            t.PrintFacts(1, prefix)
            env.Targets.Push(t)
        case j := <- env.JobCh:
            env.Jobqueue.Push(&j)
        default:
            if env.Jobqueue.Done() && len(env.FactCh) == 0 && len(env.JobCh) == 0 {
                log.Logf(0, "%vAll jobs completed.\n", prefix)
                return
            }
        }
    }
}

func main() {
    flag.Parse()
    
    if *flagURL == "" && *flagEP == "" {
        log.Err("No target was given! Please use either -url or -ep\n")
        flag.PrintDefaults()
        return
    }

    // parse Config
    // For now...
    cfg := env.DefaultConfig()
    cfg.NumWorkers = 1
    cfg.ClientCfg.Slowdown = 5*time.Second
    cfg.Seclists = "/Users/jbursey/Documents/SecLists/"

    Greeting()

    jq := &job.JobQueue{}
    jq.Init()

    httpclient := httpclient.New(cfg.ClientCfg)

    env := &env.Env{
        Jobqueue: *jq,
        Cfg:      *cfg,
        Client:   *httpclient,
        JobCh:    make(chan job.Job, 10),
        FactCh:   make(chan fact.Target, 10),
    }

    // Spawn the Worker Threads
    for i := uint(0); i < env.Cfg.NumWorkers; i++ {
        go worker.Run(i, env)
    }

    // Define the top-level target
    firstTarget := &fact.Target{
        Port:  *flagPort,
        Facts: make(map[fact.FactKey]fact.FactValue),
    }
    if *flagEP != "" {
        firstTarget.Url = *flagEP
        firstTarget.Type = fact.TargetEp
    } else {
        firstTarget.Url = *flagURL
        firstTarget.Type = fact.TargetDomain
    }
    env.Targets.Push(*firstTarget)

    // Create the first discovery job
    firstJob := &job.Job{
        Mode:     job.ModeDiscovery,
        Action:   worker.ActionCheckAlive,
        Priority: 5,
        Target:   firstTarget.Key(),
    }

    env.Jobqueue.Push(firstJob)

    // TODO: Print debug information
    Loop(env)

    // TODO: Print Findings
    env.Targets.PrintFindings()
}


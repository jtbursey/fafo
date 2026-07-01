// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"
    //"time"

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
    log.Log(0, "                                                        |/\n")
    log.Log(0, "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n\n\n")
}

func Loop(env env.Env) {
    for ;; {
        select {
        case <- env.FactCh:
            log.Log(0, "[Manager]: New fact")
        }
    }
}

func main() {
    flag.Parse()
    
    if *flagURL == "" && *flagEP == "" {
        log.Log(0, "No target was given! Please use either -url or -ep\n")
        flag.PrintDefaults()
        return
    }

    // parse Config
    // For now...
    cfg := &env.Config{
        NumWorkers: 1,
    }

    Greeting()

    jq := &job.JobQueue{}
    jq.Init()

    httpclient := &httpclient.HttpClient{

    }

    env := &env.Env{
        Jobqueue: *jq,
        Cfg:      *cfg,
        Client:   *httpclient,
        CorpusCh: make(chan string, 10),
        JobCh:    make(chan job.Job, 10),
        FactCh:   make(chan fact.Fact, 10),
    }

    // Spawn the Worker Threads
    for i := uint(0); i < env.Cfg.NumWorkers; i++ {
        go worker.Run(i, env)
    }

    // Define the top-level target
    firstTarget := &fact.Target{
        Port: *flagPort,
    }
    firstTarget.Default()
    if *flagEP != "" {
        firstTarget.Url = *flagEP
        firstTarget.Type = fact.TargetEp
    } else {
        firstTarget.Url = *flagURL
        firstTarget.Type = fact.TargetDomain
    }
    env.PushTarget(*firstTarget)

    // Create the first discovery job
    firstJob := &job.Job{
        Mode:     job.ModeDiscovery,
        Priority: 5,
        Target:   firstTarget.Url,
    }

    env.Jobqueue.Push(firstJob)

    Loop(*env)
}


// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"
    "time"

    "fafo/pkg/env"
    "fafo/pkg/job"
    "fafo/pkg/log"
    "fafo/pkg/worker"
)

// Define the command line args here
var (
    flagURL = flag.String("url", "", "The base `URL` (domain) to hit")
    flagEP = flag.String("ep", "", "The specific `Endpoint` to hit (overrides URL)")
    flagPort = flag.Uint("p", 443, "The `Port` on which to scan")
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
        time.Sleep(5 * time.Second)
        select {
        case newfact := <- env.FactCh:
            log.Logf(0, "[Manager]: New fact: %v", newfact)
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
    env := &env.Env{
        Jobqueue: *jq,
        Cfg:      *cfg,
        CorpusCh: make(chan string, 10),
        JobCh:    make(chan job.Job, 10),
        FactCh:   make(chan string, 10),
    }

    // Spawn the Worker Threads
    for i := uint(0); i < env.Cfg.NumWorkers; i++ {
        go worker.Run(i, env)
    }

    // Create the first discovery job
    firstJob := &job.Job{
        Mode:     job.ModeDiscovery,
        Priority: 5,
        Target:   *flagURL,
    }

    env.Jobqueue.Push(firstJob)

    // Enter Loop
    Loop(*env)
}


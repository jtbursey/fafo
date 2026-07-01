// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"

    "time"

    "fafo/pkg/worker"
    "fafo/pkg/log"
)

const (
    numWorkers uint = 1
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

func Loop() {
    for ;; {
        time.Sleep(100 * time.Second)
    }
}

func main() {
    flag.Parse()
    
    if *flagURL == "" && *flagEP == "" {
        log.Log(0, "No target was given! Please use either -url or -ep\n")
        flag.PrintDefaults()
        return
    }

    Greeting()

    // Spawn the Worker Threads
    for i := uint(0); i < numWorkers; i++ {
        go worker.Run()
    }

    // Create the first discovery job

    // Enter Loop
    Loop()
}


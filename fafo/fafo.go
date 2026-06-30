// Joseph Bursey <jbursey@tevora.com>

package main

import (
    "flag"
    "fmt"
    "sync"
)

// Define the command line args here
var (
    flagURL = flag.String("url", "", "The base `URL` (domain) to hit")
    flagEP = flag.String("ep", "", "The specific `Endpoint` to hit (overrides URL)")
    flagPort = flag.Uint("p", 443, "The `Port` on which to scan")
)

func main() {
    flag.Parse()
    fmt.Println("Hello %s", flagURL)
}


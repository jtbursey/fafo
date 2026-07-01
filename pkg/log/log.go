// Joseph Bursey <jbursey@tevora.com>

package log

import (
    "flag"
    "fmt"
    "sync"
)

var (
    flagVerb = flag.Int("v", 0, "The `verbosity` level (0-10)")
    mtx sync.Mutex
)

func Verb(v int) bool {
    return v <= *flagVerb
}

func Log(v int, msg string) {
    Logf(v, "%v", msg)
}

func Logf(v int, msg string, args ...any) {
    if !Verb(v) {
        return
    }

    mtx.Lock()
    defer mtx.Unlock()
    fmt.Printf(msg, args...)
}

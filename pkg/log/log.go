// Joseph Bursey <jbursey@tevora.com>

package log

import (
    "flag"
    "fmt"
    "sync"

    "fafo/pkg/pretty"
)

var (
    flagVerb = flag.Int("v", 10, "The `verbosity` level (0-10)") // Set for debugging while developing
    mtx sync.Mutex
)

// Verbosity:
// 0:  Print only positive responses
// 1:  Print Findings
// 2:  Print negative responses
// 3:  Print worker transitions (and labels)
// 4:  
// 5:  
// 6:  
// 7:  Print http Client
// 8:  
// 9:  
// 10: Print all debug information

func Verb(v int) bool {
    return v <= *flagVerb
}

func Logf(v int, msg string, args ...any) {
    if !Verb(v) {
        return
    }

    mtx.Lock()
    defer mtx.Unlock()
    fmt.Printf(msg, args...)
}

func Log(v int, msg string) {
    Logf(v, "%v", msg)
}

func Errf(msg string, args ...any) {
    Logf(0, fmt.Sprintf("%v: %v", pretty.Orange("Error"), msg), args...)
}

func Err(msg string) {
    Errf("%v", msg)
}

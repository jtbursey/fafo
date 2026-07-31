// Joseph Bursey <jbursey@tevora.com>

package log

import (
    "flag"
    "fmt"
    "sync"

    "fafo/pkg/pretty"
)

// Verbosity:
const (
    V0  int = 0         // 0:  Print only positive responses (have findings)
    V1  int = 1         // 1:  Print Findings, non-finding 404 responses
    V2  int = 2         // 2:  Print negative all responses
    V3  int = 3         // 3:  Print worker transitions (and prefixes)
    V4  int = 4         // 4:  
    V5  int = 5         // 5:  
    V6  int = 6         // 6:  
    V7  int = 7         // 7:  Verbose worker logging
    V8  int = 8         // 8:  Print Client logs
    V9  int = 9         // 9:  
    V10 int = 10        // 10: Print all debug information
)

const (
    VError  int = V0
    VWarn   int = V2
    VPrefix int = V3
)

var (
    flagVerb = flag.Int("v", V0, "The `verbosity` level (0-10)")
    mtx sync.Mutex
)

func SetVerb(v int) {
    *flagVerb = v
}

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
    Logf(V0, fmt.Sprintf("%v: %v", pretty.Orange("Error"), msg), args...)
}

func Err(msg string) {
    Errf("%v\n", msg)
}

func Greeting(msg string) {
// ______                ___                            _    ______ _           _   _____       _
// |  ___|              / _ \                          | |   |  ___(_)         | | |  _  |     | |
// | |_ _   _ ________ / /_\ \_ __ ___  _   _ _ __   __| |   | |_   _ _ __   __| | | | | |_   _| |_
// |  _| | | |_  /_  / |  _  | '__/ _ \| | | | '_ \ / _` |   |  _| | | '_ \ / _` | | | | | | | | __|
// | | | |_| |/ / / /  | | | | | | (_) | |_| | | | | (_| |_  | |   | | | | | (_| | \ \_/ / |_| | |_
// \_|  \__,_/___/___| \_| |_/_|  \___/ \__,_|_| |_|\__,_( ) \_|   |_|_| |_|\__,_|  \___/ \__,_|\__|
//                                                       |/

    Log(V0, "\n\n=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n")
    Log(V0, "  ______                ___                            _    ______ _           _   _____       _\n")
    Log(V0, "  |  ___|              / _ \\                          | |   |  ___(_)         | | |  _  |     | |\n")
    Log(V0, "  | |_ _   _ ________ / /_\\ \\_ __ ___  _   _ _ __   __| |   | |_   _ _ __   __| | | | | |_   _| |_\n")
    Log(V0, "  |  _| | | |_  /_  / |  _  | '__/ _ \\| | | | '_ \\ / _` |   |  _| | | '_ \\ / _` | | | | | | | | __|\n")
    Log(V0, "  | | | |_| |/ / / /  | | | | | | (_) | |_| | | | | (_| |_  | |   | | | | | (_| | \\ \\_/ / |_| | |_\n")
    Log(V0, "  \\_|  \\__,_/___/___| \\_| |_/_|  \\___/ \\__,_|_| |_|\\__,_( ) \\_|   |_|_| |_|\\__,_|  \\___/ \\__,_|\\__|\n")
    Logf(V0, "   %-53v|/\n", msg)
    Log(V0, "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n\n")
}

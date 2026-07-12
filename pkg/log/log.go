// Joseph Bursey <jbursey@tevora.com>

package log

import (
    "flag"
    "fmt"
    "sync"

    "fafo/pkg/pretty"
)

var (
    flagVerb = flag.Int("v", 0, "The `verbosity` level (0-10)")
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
// 7:  Print Client logs
// 8:  
// 9:  
// 10: Print all debug information

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
    Logf(0, fmt.Sprintf("%v: %v", pretty.Orange("Error"), msg), args...)
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

    Log(0, "\n\n=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n")
    Log(0, "  ______                ___                            _    ______ _           _   _____       _\n")
    Log(0, "  |  ___|              / _ \\                          | |   |  ___(_)         | | |  _  |     | |\n")
    Log(0, "  | |_ _   _ ________ / /_\\ \\_ __ ___  _   _ _ __   __| |   | |_   _ _ __   __| | | | | |_   _| |_\n")
    Log(0, "  |  _| | | |_  /_  / |  _  | '__/ _ \\| | | | '_ \\ / _` |   |  _| | | '_ \\ / _` | | | | | | | | __|\n")
    Log(0, "  | | | |_| |/ / / /  | | | | | | (_) | |_| | | | | (_| |_  | |   | | | | | (_| | \\ \\_/ / |_| | |_\n")
    Log(0, "  \\_|  \\__,_/___/___| \\_| |_/_|  \\___/ \\__,_|_| |_|\\__,_( ) \\_|   |_|_| |_|\\__,_|  \\___/ \\__,_|\\__|\n")
    Logf(0, "   %-53v|/\n", msg)
    Log(0, "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n\n")
}

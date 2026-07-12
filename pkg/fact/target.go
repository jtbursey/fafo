// Joseph Bursey <jbursey@tevora.com>

package fact

import (
    "fmt"
    "strings"
    "sync"

    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

type TargetType string

const (
    TargetDomain TargetType = "Domain"      // Specifically for the base domain
    TargetPath   TargetType = "Path"        // Some path from the base domain
    TargetEp     TargetType = "Endpoint"    // A specific endpoint (like index.html)
)

type Target struct {
    Url     string
    Port    int
    Facts   map[FactKey]FactValue           // The information we have learned
}

type TargetMap struct {
    tm  map[string]Target
    mtx sync.Mutex
}

func (tgt *Target) Key() string {
    return fmt.Sprintf("%v:%v", tgt.Url, tgt.Port)
}

func (tm *TargetMap) Pull(key string) *Target {
    tm.mtx.Lock()
    defer tm.mtx.Unlock()
    if tm.tm == nil {
        return nil
    }

    if ret, ok := tm.tm[key]; ok {
        return &ret
    }
    return nil
}

// Lock must already be held!
func (tm *TargetMap) mergeTarget(new Target) {
    old := tm.tm[new.Key()]

    for key, value := range new.Facts {
        if _, ok := old.Facts[key]; !ok {
            old.Facts[key] = value
        }

        switch {
        case key == IsAlive:
            if old.Facts[key] == True && value == False {
                old.Facts[HasDied] = True
            } else if value == True {
                old.Facts[key] = True
            }
        default:
            old.Facts[key] = value
        }
    }
}

func (tm *TargetMap) Push(target Target) {
    tm.mtx.Lock()
    defer tm.mtx.Unlock()
    if tm.tm == nil {
        tm.tm = make(map[string]Target)
    }

    _, ok := tm.tm[target.Key()]
    if ok {
        tm.mergeTarget(target)
    } else {
        tm.tm[target.Key()] = target
    }
}

func (tm *TargetMap) PrintFindings() {
    log.Log(0, "\nFindings:\n")
    for _, tgt := range tm.tm {
        log.Logf(0, "Target: %v\n", tgt.Url)
        for key, value := range tgt.Facts {
            log.Logf(0, "    [%v: %v]\n", key, value)
        }
    }
}

func (tgt *Target) IsPath() bool {
    return strings.HasSuffix(tgt.Url, "/")
}

func (tgt *Target) PrintFacts(v int, prefix string) {
    for key, value := range tgt.Facts {
        log.Logf(v, "%v%*s [%v: %v]\n", prefix, pretty.UrlWidth, tgt.Url, key, value)
    }
}

func UrlAppend(url string, newBit string) string {
    sep := ""
    if len(url) > 0 && !strings.HasSuffix(url, "/") {
        sep = "/"
    }
    return fmt.Sprintf("%v%v%v", url, sep, newBit)
}

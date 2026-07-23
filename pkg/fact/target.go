// Joseph Bursey <jbursey@tevora.com>

package fact

import (
    "fmt"
    "net/url"
    "slices"
    "strings"
    "sync"

    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

type Target struct {
    Url     *url.URL
    Facts   map[FactKey][]FactValue           // The information we have learned
}

type TargetMap struct {
    tm  map[string]Target
    mtx sync.Mutex
}

func ChopSlash(origin string) string {
    return strings.TrimSuffix(origin, "/")
}

func (tgt *Target) Key() string {
    return tgt.Url.String()
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

    for key, values := range new.Facts {
        if _, ok := old.Facts[key]; !ok {
            old.Facts[key] = values
            continue
        }

        switch key {
        case IsAlive:
            if old.Facts[key][0] == True && values[0] == False {
                old.Facts[HasDied] = []FactValue{True}
            } else if values[0] == True {
                old.Facts[key] = []FactValue{True}
            }
        case Redirects:
            old.AppendUniqueValues(key, values)
        case Path:
            old.AppendUniqueValues(key, values)
        default:
            old.Facts[key] = values
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

func (tm *TargetMap) PrettyFinding(key FactKey, values []FactValue, space int) string {
    prettyKey := string(key)
    originlen := len(prettyKey)
    switch key {
    case "Redirects":
        prettyKey = pretty.Yellow(prettyKey)
    case "Login":
        prettyKey = pretty.Green(prettyKey)
    default:
        prettyKey = pretty.Blue(prettyKey)
    }
    output := fmt.Sprintf("    | %v: %v", prettyKey, values[0])
    for _, v := range values[1:] {
        output += fmt.Sprintf("\n%*s%v", -8-originlen, "    |", v)
    }
    return output
}

func (tm *TargetMap) PrintFindings() {
    log.Log(0, "\nFindings:\n")
    for _, tgt := range tm.tm {
        space := tgt.LongestKey()
        log.Logf(0, "Target: %v\n", tgt.Url.String())
        for key, values := range tgt.Facts {
            log.Logf(0, "%v\n", tm.PrettyFinding(key, values, space))
        }
    }
}

func (tgt *Target) AppendUniqueValues(key FactKey, values []FactValue) {
    for _, v := range values {
        if !slices.Contains(tgt.Facts[key], v) {
            tgt.Facts[key] = append(tgt.Facts[key], v)
        }
    }
}

func (tgt *Target) LongestKey() int {
    best := 0
    for key, _ := range tgt.Facts {
        if len(key) > best {
            best = len(key)
        }
    }
    return best
}

func (tgt *Target) MyUrl() string {
    url := fmt.Sprintf("%v://%v", tgt.Url.Scheme, tgt.Url.Hostname())
    if tgt.Url.Port() != "" {
        url = fmt.Sprintf("%v:%v", url, tgt.Url.Port())
    }
    return url
}

func (tgt *Target) PrintFacts(v int, prefix string) {
    for key, values := range tgt.Facts {
        log.Logf(v, "%v%*s [%v: %v]\n", prefix, pretty.UrlWidth, tgt.Url.String(), key, values)
    }
}

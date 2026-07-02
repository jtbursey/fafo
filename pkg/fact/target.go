// Joseph Bursey <jbursey@tevora.com>

package fact

import (
	"fmt"
	"sync"
)

type TargetType string

const (
	TargetDomain TargetType = "domain"		// Specifically for the base domain
	TargetPath   TargetType = "path"		// Some path from the base domain
	TargetEp     TargetType = "endpoint"	// A specific endpoint (like index.html)
)

type Target struct {
	Url     string
	Type    TargetType
	Port    int
	IsAlive int
	Exists  int
}

type TargetMap struct {
	tm  map[string]Target
	mtx sync.Mutex
}

func (tgt *Target) Default() {
	tgt.IsAlive = -1
	tgt.Exists = -1
}

func (tgt *Target) Key() string {
	return fmt.Sprintf("%v:%v", tgt.Url, tgt.Port)
}

func FromFact(f Fact) Target {
	tgt := Target{
		Url:  f.Url,
		Port: f.Port,
	}

	tgt.Default()
	for key, value := range f.Novel {
		switch {
		case key == IsAlive:
			tgt.IsAlive = FactBool(value)
		case key == Exists:
			tgt.Exists = FactBool(value)
		}
	}

	return tgt
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

	if new.IsAlive >= 0 && old.IsAlive <= 0 {
		old.IsAlive = new.IsAlive
	}
	if new.Exists >= 0 {
		old.Exists = new.Exists
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

func (tm *TargetMap) PushFact(fact Fact) {
	tm.Push(FromFact(fact))
}

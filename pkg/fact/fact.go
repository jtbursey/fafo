// Joseph Bursey <jbursey@tevora.com>

package fact

import (
	"fmt"

	"fafo/pkg/log"
)

type FactKey string
type FactValue string

const (
	IsAlive     FactKey = "IsAlive"
	Exists      FactKey = "Exists"

	FactTrue    FactValue = "true"
	FactFalse   FactValue = "false"
)

// Should be easily merged with Target
type Fact struct {
	Url     string
	Port    int
	Novel   map[FactKey]FactValue // The new findings to be merged
}

func FactBool(v FactValue) int {
	if v == FactTrue {
		return 1
	} else if v == FactFalse {
		return 0
	}
	return -1
}

func (f Fact) PrintNovel(v int, prefix string) {
	for key, value := range f.Novel {
		log.Log(v, fmt.Sprintf("%v%-50v [%v: %v]\n", prefix, f.Url, key, value))
	}
}

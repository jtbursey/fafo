// Joseph Bursey  <jbursey@tevora.com>

package worker

import (
	"time"

	"fafo/pkg/log"
)

const (
	StatusIdle    string = "idle"
	StatusWorking string = "working"
)

func Loop() {
	for ;; {
		log.Log(0, "I exist!\n")
		time.Sleep(5 * time.Second)
	}
}

func Run() {
	Loop()
}

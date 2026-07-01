// Joseph Bursey <jbursey@tevora.com>

package job

const (
	ModeDiscovery String = "discovery"
	ModeFuzzy     String = "fuzzy"
	ModeAttack    String = "attack"
)

type Job struct {
	data string
}


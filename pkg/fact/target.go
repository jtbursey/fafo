// Joseph Bursey <jbursey@tevora.com>

package fact

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
}

func (tgt *Target) Default() {
	tgt.IsAlive = -1
}

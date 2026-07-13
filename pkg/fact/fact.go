// Joseph Bursey <jbursey@tevora.com>

package fact

type FactKey string
type FactValue string

const (
	IsAlive FactKey = "IsAlive"
	HasDied FactKey = "HasDied"
	Exists  FactKey = "Exists"

	True  FactValue = "true"
	False FactValue = "false"
)

func FactBool(v FactValue) int {
	switch v {
	case True:
		return 1
	case False:
		return 0
	}
	return -1
}

// Joseph Bursey <jbursey@tevora.com>

package fact

type FactKey string
type FactValue string

const (
    IsAlive     FactKey = "IsAlive"
    HasDied     FactKey = "HasDied"
    Exists      FactKey = "Exists"

    True        FactValue = "true"
    False       FactValue = "false"
)

func FactBool(v FactValue) int {
    if v == True {
        return 1
    } else if v == False {
        return 0
    }
    return -1
}

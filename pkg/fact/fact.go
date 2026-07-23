// Joseph Bursey <jbursey@tevora.com>

package fact

// import (
//     "fmt"
// )

type FactKey string
type FactValue string

const (
    IsAlive   FactKey = "IsAlive"
    HasDied   FactKey = "HasDied"
    Exists    FactKey = "Exists"
    Redirects FactKey = "Redirects"
    Path      FactKey = "Path"
)

var (
    True    FactValue = "true"
    False   FactValue = "false"
)

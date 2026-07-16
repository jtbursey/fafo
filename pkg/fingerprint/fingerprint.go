// Joseph Bursey <jbursey@tevora.com>

package fingerprint

import (
    "fmt"
    "net/http"
    "slices"
    "strings"

    "fafo/pkg/config"
    "fafo/pkg/fact"
    "fafo/pkg/job"
)

type Field string
type ConditionType string

const (
    FieldStatusCode    Field = "StatusCode"
    FieldUrl           Field = "Url"
    FieldFuzzRecursive Field = "FuzzRecursive"
    FieldHdrLocation   Field = "HdrLocation"
    FieldHdrAllow      Field = "HdrAllow"
    FieldTautology     Field = "Tautology"

    Contains           ConditionType = "Contains"
    OneOf              ConditionType = "OneOf"
    Equals             ConditionType = "Equals"
    NonEmpty           ConditionType = "NonEmpty"
)

// Field, Condition Value(s)
type Condition struct {
    Field     Field                                 `json:"Field"`
    Condition ConditionType                         `json:"Condition"`
    Values    []string                              `json:"Values"`
}

type Fingerprint []Condition

type FactConditionPair struct {
    Fingerprint Fingerprint                         `json:"Conditions"`
    FactPair    map[fact.FactKey]fact.FactValue     `json:"Facts"`
}

type JobConditionPair struct {
    Fingerprint Fingerprint                         `json:"Conditions"`
    Jobs        []job.Job                           `json:"Jobs"`
}

func (f Field) Get(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) (string, error) {
    switch f {
    case FieldStatusCode:
        return fmt.Sprintf("%v", resp.StatusCode), nil
    case FieldUrl:
        return fmt.Sprintf("%v", req.URL.String()), nil // Does this want response or request?
    case FieldFuzzRecursive:
        return fmt.Sprintf("%v", cfg.FuzzRecursive), nil
    case FieldHdrLocation:
        return resp.Header["Location"][0], nil
    case FieldHdrAllow:
        return strings.Join(resp.Header["Allow"], ","), nil
    case FieldTautology:
        return "true", nil
    default:
        return "", fmt.Errorf("Tried to Get unimplemented Field %v", f)
    }
}

func (c *Condition) Validate() bool {
    switch c.Condition {
    case Contains:
        return len(c.Values) == 1
    case Equals:
        return len(c.Values) == 1
    }
    return len(c.Values) > 0
}

func (c *Condition) doCompare(field string) bool {
    switch c.Condition {
    case OneOf:
        return slices.Contains(c.Values, field)
    case Equals:
        return field == c.Values[0]
    case Contains:
        return strings.Contains(field, c.Values[0])
    case NonEmpty:
        return len(field) > 0
    default:
        return false
    }
    return false
}

func (c *Condition) Evaluate(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) (bool, error) {
    field, err := c.Field.Get(resp, req, base, cfg)
    if err != nil {
        return false, err
    }
    return c.doCompare(field), nil
}

func (f *Fingerprint) Evaluate(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) (bool, error) {
    for _, c := range *f {
        res, err := c.Evaluate(resp, req, base, cfg)
        if !res {
            return false, err
        }
    }
    return true, nil
}

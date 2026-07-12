// Joseph Bursey <jbursey@tevora.com>

package fingerprint

import (
    "fmt"
    "net/http"
    "slices"

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

    Contains           ConditionType = "Contains"
    OneOf              ConditionType = "OneOf"
    Equals             ConditionType = "Equals"
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

func (c *Condition) Validate() bool {
    switch c.Condition {
    case Contains:
        return len(c.Values) == 1
    case Equals:
        return len(c.Values) == 1
    }
    return len(c.Values) > 0
}

func (c *Condition) getField(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) string {
    switch c.Field {
    case FieldStatusCode:
        return fmt.Sprintf("%v", resp.StatusCode)
    case FieldUrl:
        return fmt.Sprintf("%v", req.URL.String())
    case FieldFuzzRecursive:
        return fmt.Sprintf("%v", cfg.FuzzRecursive)
    default:
        return ""
    }
}

func (c *Condition) doCompare(field string) bool {
    switch c.Condition {
    case OneOf:
        return slices.Contains(c.Values, field)
    case Equals:
        return field == c.Values[0]
    default:
        return false
    }
}

func (c *Condition) Evaluate(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) bool {
    field := c.getField(resp, req, base, cfg)
    return c.doCompare(field)
}

func (f *Fingerprint) Evaluate(resp *http.Response, req *http.Request, base *fact.Target, cfg *config.Config) bool {
    for _, c := range *f {
        if !c.Evaluate(resp, req, base, cfg) {
            return false
        }
    }
    return true
}

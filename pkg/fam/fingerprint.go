// Joseph Bursey <jbursey@tevora.com>

package fam

import (
    "fmt"
    "net/http"
    "slices"

    "fafo/pkg/env"
    "fafo/pkg/fact"
    "fafo/pkg/job"
)

type Field string
type ConditionType string

const (
    FieldStatusCode    Field = "StatusCode"
    FieldUrl           Field = "Url"
    FieldFuzzRecursive Field = "FuzzRecursive"
    FieldTargetType    Field = "TargetType"

    Contains           ConditionType = "Contains"
    OneOf              ConditionType = "OneOf"
    Equals             ConditionType = "Equals"
)

// Field, Condition Type, Value(s)
type Condition struct {
    Field     Field                // The Field of request or response to check
    Condition ConditionType        // The condition
    Values    []string             // The values to check against
}

type Fingerprint []Condition

type FactConditionPair struct {
    Fingerprint Fingerprint                            // Each condition in the fingerprint must be true
    FactPair    map[fact.FactKey]fact.FactValue        // The resulting fact that is learned
}

type JobConditionPair struct {
    Fingerprint Fingerprint
    Jobs        []job.Job
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

func (c *Condition) getField(resp *http.Response, req *FamRequest, base *fact.Target, env *env.Env) string {
    switch c.Field {
    case FieldStatusCode:
        return fmt.Sprintf("%v", resp.StatusCode)
    case FieldUrl:
        return fmt.Sprintf("%v", req.Req.URL.String())
    case FieldFuzzRecursive:
        return fmt.Sprintf("%v", env.Cfg.FuzzRecursive)
    case FieldTargetType:
        return fmt.Sprintf("%v", req.Type)
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

func (c *Condition) Evaluate(resp *http.Response, req *FamRequest, base *fact.Target, env *env.Env) bool {
    field := c.getField(resp, req, base, env)
    return c.doCompare(field)
}

func (f *Fingerprint) Evaluate(resp *http.Response, req *FamRequest, base *fact.Target, env *env.Env) bool {
    for _, c := range *f {
        if !c.Evaluate(resp, req, base, env) {
            return false
        }
    }
    return true
}

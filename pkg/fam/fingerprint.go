// Joseph Bursey <jbursey@tevora.com>

package fam

import (
	"net/http"

	"fafo/pkg/fact"
	"fafo/pkg/job"
)

type Field string
type ConditionType string

const (
	FieldRespCode Field = "FieldCode"
	FieldUrl      Field = "FieldUrl"

	Contains ConditionType = "Contains"
	OneOf    ConditionType = "OneOf"
	Equals   ConditionType = "Equals"
)

// Field, Condition Type, Value(s)
type Condition struct {
	Field     Field				// The Field of request or response to check
	Condition ConditionType		// The condition
	Values    []string			// The values to check against
}

type Fingerprint []Condition

type FactConditionPair struct {
	Fingerprint Fingerprint							// Each condition in the fingerprint must be true
	FactPair    map[fact.FactKey]fact.FactValue		// The resulting fact that is learned
}

type JobConditionPair struct {
	Fingerprint Fingerprint
	Job         job.Job
}

func (c *Condition) Validate() bool {
	return len(c.Values) > 0
}

func (f *Fingerprint) Evaluate(resp *http.Response, req *FamRequest, base *fact.Target) bool {
	return false
}

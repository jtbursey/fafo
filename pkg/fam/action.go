// Joseph Bursey <jbursey@tevora.com>

package fam

import (
    "fafo/pkg/fact"
    "fafo/pkg/job"
)

type Payload struct {
    Id      string                // The string to replace in the template
    Type    fact.TargetType        // What does this payload turn the target into?
    Pl      string                // The actual payload
}

type PayloadSet struct {
    Id   string
    Type fact.TargetType
    File string
    List []Payload
}

// Describes how to make the request
type RequestTemplate struct {
    Method string                // The Method to use (i.e. "GET, PUT, FUZZ")
    Url    string                // The Url Template (i.e. "BASE/FUZZ")
    //Header *HeaderTemplate
    //Body   *BodyTemplate
}

// Job conditionals and Fact conditionals
    // If X, set key/value in fact
    // If Y, push a job (mode, action, prio, target)
type ResponseAction struct {
    // Body handler here
    Factcond []FactConditionPair
    Jobcond  []JobConditionPair
}

type Action struct {
    Id      job.Action
    Pylds   *PayloadSet
    Reqt    *RequestTemplate
    RespAct *ResponseAction
}

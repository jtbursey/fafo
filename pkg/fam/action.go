// Joseph Bursey <jbursey@tevora.com>

package fam

import (
    "fafo/pkg/fact"
    "fafo/pkg/job"
)

type Payload struct {
    Id      string                  // The string to replace in the template
    Type    fact.TargetType         // What does this payload turn the target into?
    Pl      string                  // The actual payload
}

type PayloadSet struct {
    Id   string                     `json:"Id"`
    Type fact.TargetType            `json:"Type"`
    File string                     `json:"File"`
    List []string                   `json:"List"`
}

// Describes how to make the request
type RequestTemplate struct {
    Method string                   `json:"Method"`
    Url    string                   `json:"UrlTemplate"`
    //Header *HeaderTemplate
    //Body   *BodyTemplate
}

// Job conditionals and Fact conditionals
    // If X, set key/value in fact
    // If Y, push a job (mode, action, prio, target)
type ResponseAction struct {
    // Body handler here
    ScrShcond Fingerprint           `json:"ScreenShotConditions"`
    Factcond  []FactConditionPair   `json:"FactConditions"`
    Jobcond   []JobConditionPair    `json:"JobConditions"`
}

type Action struct {
    Id      job.Action              `json:"Id"`
    Pylds   *PayloadSet             `json:"PayloadSet"`
    Reqt    *RequestTemplate        `json:"RequestTemplate"`
    RespAct *ResponseAction         `json:"ResponseAction"`
}

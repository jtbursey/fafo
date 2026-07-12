// Joseph Bursey <jbursey@tevora.com>

package action

import (
    "encoding/json"
    "fmt"
    "os"

    "fafo/pkg/fs"
    "fafo/pkg/fingerprint"
    "fafo/pkg/job"
)

type Payload struct {
    Id      string                  // The string to replace in the template
    Pl      string                  // The actual payload
}

type PayloadSet struct {
    Id   string                               `json:"Id"`
    File string                               `json:"File"`
    List []string                             `json:"List"`
}

// Describes how to make the request
type RequestTemplate struct {
    Method string                             `json:"Method"`
    Url    string                             `json:"UrlTemplate"`
    //Header *HeaderTemplate
    //Body   *BodyTemplate
}

// Job conditionals and Fact conditionals
    // If X, set key/value in fact
    // If Y, push a job (mode, action, prio, target)
type ResponseAction struct {
    // Body handler here
    ScrShcond fingerprint.Fingerprint         `json:"ScreenShotConditions"`
    Factcond  []fingerprint.FactConditionPair `json:"FactConditions"`
    Jobcond   []fingerprint.JobConditionPair  `json:"JobConditions"`
}

type Action struct {
    Id      string                            `json:"Id"`
    Mode    job.WorkerMode                    `json:"Mode"`
    Pylds   *PayloadSet                       `json:"PayloadSet"`
    Reqt    *RequestTemplate                  `json:"RequestTemplate"`
    RespAct *ResponseAction                   `json:"ResponseAction"`
}

func Parse(filename string) (*Action, error) {
    if !fs.Exists(filename) {
        err := fmt.Errorf("%v does not exist", filename)
        return nil, err
    }

    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    act := &Action{}

    err = json.Unmarshal(data, act)
    if err != nil {
        return nil, err
    }

    return act, nil
}

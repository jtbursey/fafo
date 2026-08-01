// Joseph Bursey <jbursey@tevora.com>

package fam

import (
    "fmt"
    "net/http"
    "net/url"
    "slices"
    "strings"

    "fafo/pkg/action"
    "fafo/pkg/config"
    "fafo/pkg/fingerprint"
)

// Fafo deals with a lot of 'data': port numbers, strings, urls, slices, etc.
// This structure aims to be a way to join all those together so they can be easily compared, printed, passed around, etc.

type Data struct {
    Base             *url.URL           // BASE Url from target
    Payloads         []action.Payload
    Request          *http.Request
    Response         *http.Response
    Config           *config.Config
}

func NewData() *Data {
    return &Data{}
}

// This series of functions takes in data structures and parses out what it needs.
// i.e. get the Allow header from a response packet
func (d *Data) TakeConfig(cfg *config.Config) {
    d.Config = cfg
}

func (d *Data) TakeRequest(req *http.Request) {
    d.Request = req
}

func (d *Data) TakeResponse(resp *http.Response) {
    d.Response = resp
}

func (d *Data) TakePayloads(pylds []action.Payload) {
    d.Payloads = pylds
}

func (d *Data) baseReplace(origin string) (string, error) {
    if d.Base != nil && strings.Contains(origin, "BASE") {
        origin = strings.ReplaceAll(origin, "BASE", d.Base.String())
    } else if d.Base == nil {
        return origin, fmt.Errorf("Base URL is not set yet")
    }
    return origin, nil
}

func (d *Data) currentReplace(origin string) (string, error) {
    if d.Request != nil && strings.Contains(origin, "CURRENT") {
        origin = strings.ReplaceAll(origin, "CURRENT", d.Request.URL.String())
    } else if d.Request == nil {
        return origin, fmt.Errorf("Request is not set yet")
    }
    return origin, nil
}

func (d *Data) payloadReplace(origin string) string {
    for _, pyld := range d.Payloads {
        if len(pyld.Id) > 0 && strings.Contains(origin, pyld.Id) {
            origin = strings.ReplaceAll(origin, pyld.Id, pyld.Pl)
        }
    }
    return origin
}

func (d *Data) fieldReplace(origin string) (string, error) {
    for _, field := range fingerprint.AllFields {
        if strings.Contains(origin, string(field)) {
            if val, err := d.StringField(field); err == nil {
                origin = strings.ReplaceAll(origin, string(field), val)
            } else {
                return origin, err
            }
        }
    }
    return origin, nil
}

func (d *Data) Replace(origin string) string {
    origin, _ = d.baseReplace(origin)
    origin, _ = d.currentReplace(origin)
    origin = d.payloadReplace(origin)
    origin, _ = d.fieldReplace(origin)
    return origin
}

func (d *Data) StringField(key fingerprint.Field) (string, error) {
    if d.Request == nil || d.Response == nil || d.Config == nil {
        return "", fmt.Errorf("Unable to get field")
    }
    return key.Get(d.Response, d.Request, d.Config)
}

func (d *Data) PrepareAppend(key string) []string {
    key, _ = d.baseReplace(key)
    key, _ = d.currentReplace(key)
    key = d.payloadReplace(key)

    if slices.Contains(fingerprint.AllFields, key) {
        // build the array and return it
    }

    return []string{key}
}

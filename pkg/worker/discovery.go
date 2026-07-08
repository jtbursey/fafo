// Joseph Bursey <jbursey@tevora.com>

// This file defines various discovery actions that can be carried out by workers

package worker

import (
	"fafo/pkg/env"
	"fafo/pkg/fact"
	"fafo/pkg/fam"
	"fafo/pkg/job"
)

const (
	ActionCheckAlive      job.Action = "CheckAlive"
	ActionFuzzDirectories job.Action = "FuzzDirectories"
	ActionFuzzFiles       job.Action = "FuzzFiles"
	ActionWordPressScan   job.Action = "WordPressScan"
)

var(
	CheckAlive = fam.Action{
		Id: ActionCheckAlive,
		Pylds: &fam.PayloadSet{
			Id:       "",
			File:     "",
			List:     nil,
			Type:     fact.TargetDomain,
		},
		Reqt: &fam.RequestTemplate{
			Method:   "GET",
			Url:      "BASE",
		},
		RespAct: &fam.ResponseAction{
			Factcond: []fam.FactConditionPair{fam.FactConditionPair{
				fam.Fingerprint{
					fam.Condition{
						Field: fam.FieldStatusCode,
						Condition: fam.OneOf,
						Values: []string {"200", "204", "301", "302", "307", "401"},
					},
				},
				map[fact.FactKey]fact.FactValue {
					fact.IsAlive: fact.True,
					fact.Exists: fact.True,
				},
			}},
			Jobcond: []fam.JobConditionPair{fam.JobConditionPair{
				fam.Fingerprint{
					fam.Condition{
						Field: fam.FieldStatusCode,
						Condition: fam.OneOf,
						Values: []string {"200", "204", "301", "302", "307", "401"},
					},
				},
				[]job.Job{
					job.Job{
						Mode: job.ModeDiscovery,
						Action: ActionFuzzFiles,
						Priority: 5,
						Target: "CURRENT",
					},
					job.Job{
						Mode: job.ModeDiscovery,
						Action: ActionFuzzDirectories,
						Priority: 5,
						Target: "CURRENT",
					},
				},
			}},
		},
	}

	FuzzFiles = fam.Action{
		Id: ActionFuzzFiles,
		Pylds: &fam.PayloadSet{
			Id:       "FUZZ",
			File:     "/Users/jbursey/Documents/SecLists/Discovery/Web-Content/raft-medium-files.txt",
			List:     nil,
			Type:     fact.TargetEp,
		},
		Reqt: &fam.RequestTemplate{
			Method:   "GET",
			Url:      "BASE/FUZZ",
		},
		RespAct: &fam.ResponseAction{
			Factcond: []fam.FactConditionPair{fam.FactConditionPair{
				fam.Fingerprint{fam.Condition{
					Field: fam.FieldStatusCode,
					Condition: fam.OneOf,
					Values: []string {"200", "204", "301", "302", "307", "401"},
				}},
				map[fact.FactKey]fact.FactValue {
					fact.Exists: fact.True,
				},
			}},
			Jobcond:  nil,
		},
	}

	FuzzDirectories = fam.Action{
		Id: ActionFuzzDirectories,
		Pylds: &fam.PayloadSet{
			Id:       "FUZZ",
			File:     "/Users/jbursey/Documents/SecLists/Discovery/Web-Content/raft-medium-directories-lowercase.txt",
			List:     nil,
			Type:     fact.TargetPath,
		},
		Reqt: &fam.RequestTemplate{
			Method:   "GET",
			Url:      "BASE/FUZZ",
		},
		RespAct: &fam.ResponseAction{
			Factcond: []fam.FactConditionPair{fam.FactConditionPair{
				fam.Fingerprint{fam.Condition{
					Field: fam.FieldStatusCode,
					Condition: fam.OneOf,
					Values: []string {"200", "204", "301", "302", "307", "401"},
				}},
				map[fact.FactKey]fact.FactValue {
					fact.Exists: fact.True,
				},
			}},
			Jobcond: []fam.JobConditionPair{fam.JobConditionPair{
				fam.Fingerprint{
					fam.Condition{
						Field: fam.FieldStatusCode,
						Condition: fam.OneOf,
						Values: []string {"200", "204", "301", "302", "307", "401"},
					},
					fam.Condition{
						Field: fam.FieldFuzzRecursive,
						Condition: fam.Equals,
						Values: []string {"true"},
					},
				},
				[]job.Job{
					job.Job{
						Mode: job.ModeDiscovery,
						Action: ActionFuzzFiles,
						Priority: 5,
						Target: "CURRENT",
					},
					job.Job{
						Mode: job.ModeDiscovery,
						Action: ActionFuzzDirectories,
						Priority: 5,
						Target: "CURRENT",
					},
				},
			}},
		},
	}

	DiscoveryDispatchMap = map[job.Action]fam.Action {
		ActionCheckAlive:      CheckAlive,
		ActionFuzzFiles:       FuzzFiles,
		ActionFuzzDirectories: FuzzDirectories,
	}
)

func (w *Worker) discoveryDispatch(j *job.Job, t *fact.Target, e *env.Env) {
	f := fam.Fam{
		Caller: w.IdString(),
	}

	if action, ok := DiscoveryDispatchMap[j.Action]; ok {
		f.Run(t, &action, e)
	} else {
		w.Logf(0, "Found unimplemented Job Action: %v\n", j.Action)
	}
}

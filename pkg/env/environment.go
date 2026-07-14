// Joseph Bursey <jbursey@tevora.com>

package env

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"fafo/pkg/action"
	"fafo/pkg/config"
	"fafo/pkg/fact"
	"fafo/pkg/fs"
	"fafo/pkg/httpclient"
	"fafo/pkg/job"
	"fafo/pkg/log"
	"fafo/pkg/pretty"
)

const (
    DefaultWorkflowDir string = "workflow/"
)

type Env struct {
    Jobqueue    job.JobQueue  // The queue of jobs for workers to pull from
    Cfg         config.Config // Extra Config (sommeday to be set by the user)
    Actions     map[string]action.Action
    FirstTarget fact.Target
    Targets     fact.TargetMap // Keep information about known targets
    Client      httpclient.HttpClient
    ScrShCh     chan http.Request // Channel for pushing screenshot requests
    JobCh       chan job.Job      // Channel for pushing more jobs (to mgr)
    FactCh      chan fact.Target  // Channel for pushing facts/results  (to mgr)
}

func (env *Env) Debug() {
    if env.FirstTarget.Url != "" {
        log.Logf(0, "%v\n", pretty.Config("Target", env.FirstTarget.Url))
    }
    env.Cfg.Debug()
    log.Log(0, "\n\n")
}

func (env *Env) parseActionsInDir(dir string) error {
    files, err := os.ReadDir(dir)
    if err != nil {
        log.Errf("Failed to read directory %v: %v\n", dir, err)
        return err
    }

    if env.Actions == nil {
        env.Actions = make(map[string]action.Action)
    }

    for _, file := range files {
        filename := filepath.Join(dir, file.Name())
        act, err := action.Parse(filename)
        if err != nil {
            log.Errf("Failed to parse json file %v: %v\n", filename, err)
            return err
        } else if act == nil {
            err = fmt.Errorf("%v was not parsed", filename)
            log.Errf("%v\n", err)
            return err
        }

        env.Actions[act.Id] = *act
    }
    return nil
}

func (env *Env) ParseActions() error {
    if err := env.parseActionsInDir(DefaultWorkflowDir + "discovery/"); err != nil {
        return err
    }

    // if err := env.parseActionsInDir(DefaultWorkflowDir+"fuzzy/"); err != nil {
    //     return err
    // }

    // if err := env.parseActionsInDir(DefaultWorkflowDir+"attack/"); err != nil {
    //     return err
    // }

    return nil
}

func (env *Env) FixPayloadFile(act *action.Action) error {
    if def, ok := config.DefaulSeclistFiles[act.Pylds.File]; ok {
        log.Logf(0, "\"%v\" has a default SecLists file. Using default: \"%v\"\n", act.Pylds.File, def)
        seclists := env.Cfg.NeedSeclists()
        env.Cfg.PayloadFiles[act.Pylds.File] = filepath.Join(seclists, def)
    } else {
        log.Logf(0, "Treating \"%v\" as a key that has not been defined.\n", act.Pylds.File)
        filename := fs.GetFileFromStdio("Path to payload file")
        env.Cfg.PayloadFiles[act.Pylds.File] = filename
    }

    // TODO: Make sure the new file exists

    return nil
}

func (env *Env) ValidateAction(act *action.Action) error {
    if act.Pylds != nil {
        if _, err := env.Cfg.GetAsFilename(act.Pylds.File); err != nil {
            log.Logf(0, "Payload file \"%v\" (%v) is not a valid file or defined key.\n", act.Pylds.File, act.Id)
            env.FixPayloadFile(act)
        }
    }

    // TODO: Make sure all the new Jobs are valid actions

    return nil
}

func (env *Env) Validate() error {
    if len(env.Actions) <= 0 {
        return fmt.Errorf("No Actions were parsed")
    }

    if env.Cfg.PayloadFiles == nil {
        env.Cfg.PayloadFiles = make(map[string]string)
    }

    for _, act := range env.Actions {
        if err := env.ValidateAction(&act); err != nil {
            return err
        }
    }

    return nil
}

// Joseph Bursey <jbursey@tevora.com>

package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"fafo/pkg/chrome"
	"fafo/pkg/config"
	"fafo/pkg/env"
	"fafo/pkg/fact"
	"fafo/pkg/fs"
	"fafo/pkg/httpclient"
	"fafo/pkg/job"
	"fafo/pkg/log"
	"fafo/pkg/pretty"
	"fafo/pkg/worker"
)

// Define the command line args here
var (
	flagURL     = flag.String("url", "", "The base `URL` (domain) to hit")
	flagEP      = flag.String("ep", "", "The specific `Endpoint` to hit (overrides URL)")
	flagPort    = flag.Int("p", 443, "The `Port` on which to scan")
	flagOut     = flag.String("o", config.DefaultFindingsDir, "The `Directory` in which to put the findings")
	flagC       = flag.String("c", "", "The `Config File` to use")
	flagConfig  = flag.String("config", config.DefaultConfigFile, "The `Config File` to use")
	flagNoScrSh = flag.Bool("disable-screenshot", false, "Disable all screenshotting functionality")
)

func Loop(env *env.Env) {
	prefix := ""
	if log.Verb(3) {
		prefix = fmt.Sprintf("%*s", pretty.PrefixWidth, "[Manager]: ")
	}
	for {
		select {
		case t := <-env.FactCh:
			t.PrintFacts(1, prefix)
			env.Targets.Push(t)
			// TODO: Output the existing URLs to findings dir
		case j := <-env.JobCh:
			env.Jobqueue.Push(j)
		default:
			if env.Jobqueue.Done() && len(env.FactCh) == 0 && len(env.JobCh) == 0 {
				log.Logf(0, "%vAll jobs completed.\n", prefix)
				return
			}
		}
	}
}

func main() {
	flag.Parse()

	// TODO: combine this into single url, then identify it automatically
	// TODO: Convert my url strings into go urls
	if *flagURL == "" && *flagEP == "" {
		log.Err("No target was given! Please use either -url or -ep\n")
		flag.PrintDefaults()
		return
	}

	cfg := config.DefaultConfig()
	if *flagC != "" {
		cfg.SelfFile = *flagC
	} else {
		cfg.SelfFile = *flagConfig
	}

	if err := cfg.Parse(); err != nil {
		return
	}
	cfg.FindingsDir = *flagOut

	log.Greeting("By Ocelot")

	jq := &job.JobQueue{}
	jq.Init()

	httpclient := httpclient.New(cfg.ClientCfg)

	env := &env.Env{
		Jobqueue: *jq,
		Cfg:      *cfg,
		Client:   *httpclient,
		ScrShCh:  make(chan http.Request, 10),
		JobCh:    make(chan job.Job, 10),
		FactCh:   make(chan fact.Target, 10),
	}

	if err := env.ParseActions(); err != nil {
		return
	}

	// Define the top-level target
	firstTarget := &fact.Target{
		Port:  *flagPort,
		Facts: make(map[fact.FactKey]fact.FactValue),
	}
	if *flagEP != "" {
		firstTarget.Url = *flagEP
	} else {
		firstTarget.Url = *flagURL
	}
	env.FirstTarget = *firstTarget
	env.Targets.Push(*firstTarget)

	// Create the first discovery job
	firstJob := job.Job{
		Mode:     job.ModeDiscovery,
		Action:   "CheckAlive",
		Priority: 5,
		Target:   firstTarget.Key(),
	}

	if err := env.Validate(); err != nil {
		log.Errf("Environment failed validation: %v\n", err)
		return
	}

	env.Debug()

	if err := fs.Mkdir(env.Cfg.FindingsDir); err != nil {
		log.Errf("Failed to create directory %v: %v\n", env.Cfg.FindingsDir, err)
		return
	}

	var chrm *chrome.Chrome
	if !*flagNoScrSh {
		env.Cfg.ScrShDir = filepath.Join(env.Cfg.FindingsDir, "screenshots")
		if err := fs.Mkdir(env.Cfg.ScrShDir); err != nil {
			log.Errf("Failed to mkdir %v: %v", env.Cfg.ScrShDir, err)
			return
		}

		chrm = chrome.NewChrome(env)
		if chrm == nil {
			log.Err("Failed to launch Chrome!")
			return
		}
		go chrm.Loop(env)
	} else {
		log.Log(0, "Screenshotting is Disabled\n")
		env.Cfg.DisableScreenShot = true
	}

	// Spawn the Worker Threads
	for i := uint(0); i < env.Cfg.NumWorkers; i++ {
		go worker.Run(i, env)
	}

	// This kicks everything off
	env.Jobqueue.Push(firstJob)
	Loop(env)

	// This doesn't actually do much does it.
	chrm.SignalDone()

	// TODO: Findings to output dir
	env.Targets.PrintFindings()
}

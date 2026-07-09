// Joseph Bursey <jbursey@tevora.com>

package chrome

import (
    "context"
    "fmt"
    "os"

    "github.com/chromedp/cdproto/network"
    "github.com/chromedp/chromedp"

    "fafo/pkg/env"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

type Chrome struct {
    inst          *instance
    browserCtx    context.Context
	browserCancel context.CancelFunc
    baseTasks     chromedp.Tasks
    doneSignal    chan bool
}

type instance struct {
    Ctx      context.Context
    Cancel   context.CancelFunc
    UserData string
}

func (inst *instance) Close() {
	inst.Cancel()
	<-inst.Ctx.Done()
	if inst.UserData != "" {
		os.RemoveAll(inst.UserData)
	}
}

func newInstance(env *env.Env) *instance {
    userDataDir, err := os.MkdirTemp("", "fafo-chrome-*")
    if err != nil {
        return nil
    }

    opts := append(
        chromedp.DefaultExecAllocatorOptions[:],
        chromedp.NoDefaultBrowserCheck,
        chromedp.NoFirstRun,
        chromedp.DisableGPU,
        chromedp.IgnoreCertErrors,
        chromedp.UserAgent(env.Cfg.ClientCfg.UserAgent),
        chromedp.Flag("disable-features", "MediaRouter,HttpsUpgrades,OptimizationHints,AutofillServerCommunication"),
        chromedp.Flag("mute-audio", true),
        chromedp.Flag("hide-scrollbars", true),
        chromedp.Flag("disable-background-timer-throttling", true),
        chromedp.Flag("disable-backgrounding-occluded-windows", true),
        chromedp.Flag("disable-renderer-backgrounding", true),
        chromedp.Flag("disable-background-networking", true),
        chromedp.Flag("disable-component-update", true),
        chromedp.Flag("disable-domain-reliability", true),
        chromedp.Flag("disable-sync", true),
        chromedp.Flag("metrics-recording-only", true),
        chromedp.Flag("no-pings", true),
        chromedp.Flag("disable-extensions", true),
        chromedp.Flag("disable-breakpad", true),
        chromedp.Flag("disable-crash-reporter", true),
        chromedp.Flag("disable-translate", true),
        chromedp.Flag("deny-permission-prompts", true),
        chromedp.Flag("https-upgrades-enabled", false),
        //chromedp.Flag("explicitly-allowed-ports", ports),
        chromedp.Flag("no-sandbox", true),
        chromedp.WindowSize(1920, 1080),
        chromedp.UserDataDir(userDataDir),
    )

    // Proxy and specific Chrome path go here
    // append(opts, chromedp.ProxyServer(Proxy))
    // append(opts, chromedp.ExecPath(Path))

    ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

    return &instance{
        Ctx:      ctx,
        Cancel:   cancel,
        UserData: userDataDir,
    }
}

func NewChrome(env *env.Env) *Chrome {
    inst := newInstance(env)
    if inst == nil {
        return nil
    }

    browserCtx, browserCancel := chromedp.NewContext(inst.Ctx)

    if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(context.Context) error { return nil })); err != nil {
		browserCancel()
		inst.Close()
		return nil
	}

    chrome := &Chrome{
        inst:          inst,
        browserCtx:    browserCtx,
        browserCancel: browserCancel,
        doneSignal:    make(chan bool, 0),
    }

    chrome.baseTasks = chromedp.Tasks{
		network.Enable(),
	}

	return chrome
}

func (c *Chrome) prefix() string {
    prefix := ""
    if log.Verb(3) {
        prefix = fmt.Sprintf("%*s", pretty.PrefixWidth, "[Chrome]: ")
    }
    return prefix
}

func (c *Chrome) Logf(v int, msg string, args ...any) {
    log.Logf(v, c.prefix()+msg, args...)
}

func (c *Chrome) Log(v int, msg string) {
    c.Logf(v, "%v", msg)
}

func (c *Chrome) Errf(msg string, args ...any) {
    log.Logf(0, fmt.Sprintf("%*s%v: %v\n", pretty.PrefixWidth, c.prefix(), pretty.Orange("Error"), msg), args...)
}

func (c *Chrome) Err(msg string) {
    c.Errf("%v", msg)
}

func (c *Chrome) ScreenShot() {
    c.Log(0, "Snap!\n")
    // Todo: Add a callback thing to grab the call semaphore from httpclient
}

func (c *Chrome) Loop(env *env.Env) {
    c.Log(7, "Chrome has Started\n")
    for {
        select {
        case <- env.ScrShCh:
            c.ScreenShot()
        default:
            if c.checkDone() { return }
        }
    }
}

func (c *Chrome) SignalDone() {
    c.doneSignal <- true
}

func (c *Chrome) checkDone() bool {
    select {
    case isdone := <- c.doneSignal:
        if isdone {
            return true
        }
    default:
        return false
    }
    return false
}

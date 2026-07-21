// Joseph Bursey <jbursey@tevora.com>

package chrome

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    "unicode"

    "github.com/chromedp/cdproto/network"
    "github.com/chromedp/cdproto/page"
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

    Extension string
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
        doneSignal:    make(chan bool),
        Extension:     env.Cfg.ScrShExt,
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
    log.Logf(0, fmt.Sprintf("%*s%v: %v", pretty.PrefixWidth, c.prefix(), pretty.Orange("Error"), msg), args...)
}

func (c *Chrome) Err(msg string) {
    c.Errf("%v\n", msg)
}

func timestamp() string {
    var stamp string
    stamp = time.Now().Format("2006-01-02 15:04:05")
    stamp = strings.ReplaceAll(stamp, "-", "")
    stamp = strings.ReplaceAll(stamp, " ", "")
    stamp = strings.ReplaceAll(stamp, ":", "")
    return stamp
}

func safeFilename(origin string) string {
    origin = strings.ReplaceAll(origin, "://", ".")
    ret := ""
    for _, c := range origin {
        if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '.' {
            ret += string(c)
        } else {
            ret += "-"
        }
    }
    return ret
}

func (c *Chrome) craftPathname(req *http.Request, env *env.Env) string {
    return filepath.Join(env.Cfg.ScrShDir, fmt.Sprintf("%v-%v.%v", timestamp(), safeFilename(req.URL.String()), c.Extension))
}

func (c *Chrome) ScreenShot(req *http.Request, env *env.Env) {
    // Borrowing here because somehthing times out if I make the context and then wait.
    env.Client.BorrowSem()
    tabCtx, tabCancel := chromedp.NewContext(c.browserCtx)
    defer tabCancel()

    navigationCtx, navigationCancel := context.WithTimeout(tabCtx, env.Cfg.ClientCfg.Timeout)
    defer navigationCancel()

    tasks := append(chromedp.Tasks{}, c.baseTasks...)

    tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
        if err := network.ClearBrowserCookies().Do(ctx); err != nil {
            c.Errf("Failed to clear Cookies: %v\n", err)
        }

        if err := chromedp.Navigate(req.URL.String()).Do(ctx); err != nil {
            return err
        }

        if err := chromedp.WaitReady("body", chromedp.ByQuery).Do(ctx); err != nil {
            return err
        }

        return nil
    }))

    var img []byte
    tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
        params := page.CaptureScreenshot().WithQuality(int64(60)).WithFormat(page.CaptureScreenshotFormat(c.Extension))

        var err error
        img, err = params.Do(ctx)
        if err != nil {
            c.Errf("Screenshot was not captured: %v\n", err)
            return err
        }

        return nil
    }))

    if err := chromedp.Run(navigationCtx, tasks); err != nil {
        c.Errf("Failed to capture Screenshot: %v\n", err)
        env.Client.ReturnSem()
        return
    }
    env.Client.ReturnSem()

    filename := c.craftPathname(req, env)
    if err := os.WriteFile(filename, img, os.FileMode(0664)); err != nil {
        c.Errf("Failed to save screenshot: %w\n", err)
        return
    }

    c.Logf(0, "%v\n", pretty.Screenshot(req.URL.String(), filename))

    // An option:
    // decoded, _, err := image.Decode(bytes.NewReader(img))
    // if err != nil {
    //     return nil, fmt.Errorf("failed to decode screenshot image: %w", err)
    // }

    // hash, err := imagehash.PerceptionHash(decoded)
    // if err != nil {
    //     return nil, fmt.Errorf("failed to calculate image perception hash: %w", err)
    // }
}

func (c *Chrome) Loop(env *env.Env) {
    c.Log(7, "Chrome has Started\n")
    for {
        select {
        case req := <-env.ScrShCh:
            c.ScreenShot(&req, env)
        default:
            if c.checkDone() {
                return
            }
        }
    }
}

func (c *Chrome) SignalDone() {
    if c == nil {
        return
    }
    c.doneSignal <- true
}

func (c *Chrome) checkDone() bool {
    select {
    case isdone := <-c.doneSignal:
        if isdone {
            return true
        }
    default:
        return false
    }
    return false
}

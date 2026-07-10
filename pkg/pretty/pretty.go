// Joseph Bursey <jbursey@tevora.com>

package pretty

import (
    "fmt"
    "net/http"
)

const (
    UrlWidth    int = -75
    PrefixWidth int = -13
    ConfigWidth int = -25
)

func Red(str string) string {
    return fmt.Sprintf("\033[31m%v\033[0m", str)
}

func Orange(str string) string {
    return fmt.Sprintf("\033[38;2;255;165;0m%v\033[0m", str)
}

func Yellow(str string) string {
    return fmt.Sprintf("\033[33m%v\033[0m", str)
}

func Green(str string) string {
    return fmt.Sprintf("\033[32m%v\033[0m", str)
}

func Blue(str string) string {
    return fmt.Sprintf("\033[34m%v\033[0m", str)
}

func ColorCode(code int) string {
    scode := fmt.Sprintf("%v", code)
    switch code {
    case 200:
        return Green(scode)
    case 302:
        return Green(scode)
    case 403:
        return Orange(scode)
    case 404:
        return Red(scode)
    case 500:
        return Yellow(scode)
    }
    return Yellow(scode)
}

func Config(label string, config any) string {
    return fmt.Sprintf("      <=> %*s%v", ConfigWidth, fmt.Sprintf("%v: ", label), config)
}

func Response(resp *http.Response, url string) string {
    return fmt.Sprintf("%*s [Status: %v]", UrlWidth, url, ColorCode(resp.StatusCode))
}

func Screenshot(url string, pathname string) string {
    return fmt.Sprintf("%*s [Screenshot: %v]", UrlWidth, url, pathname)
}

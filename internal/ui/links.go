package ui

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// urlRegex matches http/https URLs in plain text. It does NOT stop at newlines
// so URLs that are hard-wrapped across textarea rows are captured in full.
// Stops at: space, tab, ESC, ), ], >, ", '
var urlRegex = regexp.MustCompile(`https?://[^ \t\x1b\)\]\>"'\n\r]*(?:[\n\r]+[^\s\x1b\)\]\>"']+)*`)

// extractURLs returns all unique URLs found in a plain-text string.
// Embedded newlines and carriage returns (from word-wrap or paste) are stripped
// from each matched URL before it is returned.
func extractURLs(text string) []string {
	matches := urlRegex.FindAllString(text, -1)
	seen := make(map[string]bool)
	var urls []string
	for _, u := range matches {
		// Strip any whitespace that snuck into the URL via hard-wrapping.
		u = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' {
				return -1 // drop
			}
			return r
		}, u)
		u = strings.TrimRight(u, ".,;:!?")
		if u != "" && !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}
	return urls
}

// wrapLinksOSC8 replaces every URL in text with an OSC 8 hyperlink sequence.
// The full (newline-stripped) URL is embedded in the escape so the terminal
// associates the entire visible text — even when it wraps — with the right URL.
func wrapLinksOSC8(text string) string {
	return urlRegex.ReplaceAllStringFunc(text, func(match string) string {
		url := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' {
				return -1
			}
			return r
		}, match)
		url = strings.TrimRight(url, ".,;:!?")
		if url == "" {
			return match
		}
		// Use the cleaned url as display text, not the raw match
		return "\x1b]8;;" + url + "\x1b\\" + url + "\x1b]8;;\x1b\\"
	})
}

// openURL opens the given URL in the system default browser.
func openURL(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		if isWSL() {
			// On WSL, explorer.exe is the most reliable way to open a URL
			// in the Windows default browser.
			cmd = "explorer.exe"
			args = []string{url}
		} else {
			cmd = "xdg-open"
			args = []string{url}
		}
	}
	return exec.Command(cmd, args...).Start()
}

// isWSL reports whether the process is running inside Windows Subsystem for Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

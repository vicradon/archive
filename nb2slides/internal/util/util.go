package util

import (
	"regexp"
	"strings"
)

// EscapeHTML escapes the 5 XML/HTML special characters.
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// StripHTML removes HTML tags
func StripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

// StripANSI removes ANSI escape sequences (used in traceback output)
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)

func StripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

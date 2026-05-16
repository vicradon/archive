package progress

import (
	"fmt"
	"strings"
)

// Print renders an in-place progress bar to stdout.
func Print(label string, done, total int) {
	pct := float64(done) / float64(total)
	barW := 40
	filled := int(pct * float64(barW))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
	fmt.Printf("\r  %s [%s] %3d%%  %d/%d", label, bar, int(pct*100), done, total)
}

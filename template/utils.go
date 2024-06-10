package template

import (
	"strings"
)

// CleanPercent removes leading and trailing spaces from a percentage string along with the percent sign.
func CleanPercent(percent string) string {
	percent = strings.ReplaceAll(percent, "%", "")
	return strings.TrimSpace(percent)
}

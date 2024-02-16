package rfc8601

import (
	"strings"
	"time"
)

// Format a duration as RFC-8601 in a simplified matter that doesn't
// include year-month-day section, but only the latter time section.
func Format(duration time.Duration) string {
	return "PT" + strings.ToUpper(duration.Truncate(time.Second).String())
}

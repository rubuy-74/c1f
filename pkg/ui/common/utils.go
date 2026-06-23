package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func RelativeTime(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d mins ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}
	return t.Format("2006-01-02")
}

func FormatDurationMs(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	}
	seconds := ms / 1000
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	minutes := seconds / 60
	seconds = math.Mod(seconds, 60)
	return fmt.Sprintf("%.0fm %.0fs", minutes, seconds)
}

// FormatStepOutput takes a JSON raw message from a step output and returns a
// nicely formatted string. If the output is a JSON string, it is unescaped.
// If the unescaped content is itself JSON, it is indented. Otherwise it is
// returned as plain text.
func FormatStepOutput(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try to unmarshal as a plain string first (handles escaped JSON strings).
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		// The output was a JSON string. Try to pretty-print if it's JSON.
		trimmed := bytes.TrimSpace([]byte(str))
		if len(trimmed) == 0 {
			return ""
		}
		if trimmed[0] == '{' || trimmed[0] == '[' {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, trimmed, "", "  "); err == nil {
				return pretty.String()
			}
		}
		return str
	}

	// Not a string; try to pretty-print as JSON object/array.
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err == nil {
		return pretty.String()
	}

	// Fallback: raw text.
	return string(raw)
}

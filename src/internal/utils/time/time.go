package time

import "time"

// GetCurrentTimestamp returns current UTC time in milliseconds.
func GetCurrentTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

// FormatISO formats millis into ISO timestamp: 2025-11-21T18:22:11Z
func FormatISO(ms int64) string {
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}

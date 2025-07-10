package utils

import (
	"log"
	"time"
)

// LogTimezone is the timezone used for all debug logs (always PST)
const LogTimezone = "America/Los_Angeles"

// FormatTimeForLogs formats time for debug logs in PST
func FormatTimeForLogs(t time.Time) string {
	loc, err := time.LoadLocation(LogTimezone)
	if err != nil {
		log.Printf("Error loading log timezone: %v, falling back to UTC", err)
		loc = time.UTC
	}
	tzTime := t.In(loc)
	return tzTime.Format("2006-01-02 15:04:05 MST")
}

// FormatTimeForDisplay formats time for user-facing messages in the configured timezone
func FormatTimeForDisplay(t time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Error loading display timezone %s: %v, falling back to %s", timezone, err, LogTimezone)
		loc, _ = time.LoadLocation(LogTimezone)
	}
	tzTime := t.In(loc)
	return tzTime.Format("2006-01-02 15:04:05 MST")
}

// GetLogLocation returns the timezone location for logs (PST)
func GetLogLocation() *time.Location {
	loc, err := time.LoadLocation(LogTimezone)
	if err != nil {
		log.Printf("Error loading log timezone: %v, falling back to UTC", err)
		return time.UTC
	}
	return loc
}

// GetDisplayLocation returns the timezone location for user-facing messages
func GetDisplayLocation(timezone string) *time.Location {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Error loading display timezone %s: %v, falling back to %s", timezone, err, LogTimezone)
		loc, _ = time.LoadLocation(LogTimezone)
	}
	return loc
}

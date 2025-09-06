package utils

import (
	"time"

	"github.com/google/uuid"
)

// Month abbreviations
var months = [...]string{
	"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
	"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
}

// GenerateID returns a UUID with a time-based prefix like "JUL25-<uuid>"
func GenerateID() string {
	now := time.Now()
	month := months[now.Month()-1]
	year := now.Format("06") // last 2 digits
	u := uuid.New()          // UUIDv4

	return month + year + "-" + u.String()
}

// GetIDPrefix returns the first 5 chars of an ID, e.g., "JUL25"
func GetIDPrefix(id string) string {
	if len(id) >= 5 {
		return id[:5]
	}
	return ""
}

package utils

import (
	"time"
)

const dateLayout = "2006-01-02"

// Today returns today's date formatted as YYYY-MM-DD.
func Today() string {
	return time.Now().Format(dateLayout)
}

// FutureDate returns a date that is `daysAhead` days from today, formatted as YYYY-MM-DD.
func FutureDate(daysAhead int) string {
	return time.Now().AddDate(0, 0, daysAhead).Format(dateLayout)
}

// CREstablishmentDate returns today + 30 days.
func CREstablishmentDate() string {
	return FutureDate(30)
}

// CRExpiryDate returns a date comfortably in the future (today + 365 days).
func CRExpiryDate() string {
	return FutureDate(365)
}

// OfferingStartAt returns today's date.
func OfferingStartAt() string {
	return Today()
}

// OfferingEndAt returns today + 60 days (well after the minimum 30-day requirement).
func OfferingEndAt() string {
	return FutureDate(60)
}

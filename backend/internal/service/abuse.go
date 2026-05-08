package service

import "time"

const maxAbuseRange = 90 * 24 * time.Hour

func normalizeAbuseRange(from time.Time, to time.Time) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	if to.IsZero() {
		to = now
	}
	if from.IsZero() {
		from = to.Add(-24 * time.Hour)
	}
	from = from.UTC()
	to = to.UTC()
	if !from.Before(to) {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	if to.Sub(from) > maxAbuseRange {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return from, to, nil
}

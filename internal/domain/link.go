package domain

import "time"

type Link struct {
	URL         string
	Tags        []string
	Filters     []string
	ID          int64
	LastUpdated time.Time
}

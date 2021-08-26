package util

import "time"

type Clock interface {
	Now() time.Time
}

type EagerClock struct{}

func (e EagerClock) Now() time.Time {
	return time.Now()
}

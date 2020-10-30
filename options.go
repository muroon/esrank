package esrank

import "time"

// Option function for option
type Option func(r *Ranking) *Ranking

// Name option for name setting
func Name(name string) Option {
	return func(r *Ranking) *Ranking {
		r.name = name
		return r
	}
}

// SetTimeMode option for time mode
func SetTimeMode(mode TimeMode) Option {
	return func(r *Ranking) *Ranking {
		r.mode = mode
		return r
	}
}

// StartTime option for start time
func StartTime(st time.Time) Option {
	return func(r *Ranking) *Ranking {
		r.startTime = st
		return r
	}
}


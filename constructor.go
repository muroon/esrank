package esrank

import "time"

type Option func(r *Ranking) *Ranking

func Name(name string) Option {
	return func(r *Ranking) *Ranking {
		r.name = name
		return r
	}
}

func SetTimeMode(mode TimeMode) Option {
	return func(r *Ranking) *Ranking {
		r.mode = mode
		return r
	}
}

func StartTime(st time.Time) Option {
	return func(r *Ranking) *Ranking {
		r.startTime = st
		return r
	}
}


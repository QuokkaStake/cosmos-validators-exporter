package types

import (
	"time"
)

type QueryInfo struct {
	Chain    string
	URL      string
	Duration time.Duration
	Success  bool
}

type Amount struct {
	Amount float64
	Denom  string
}

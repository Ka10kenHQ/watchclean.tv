package scraper

import "time"

const (
	MaxMoviePages     = 332
	MaxShowPages      = 38
	MovieParallelism  = 1
	ShowParallelism   = 1
)

const (
	RequestTimeout  = 30 * time.Second
	RequestDelay    = 10 * time.Second
	RandomDelay     = 2 * time.Second
)

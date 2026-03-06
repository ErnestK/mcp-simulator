package server

import "time"

type Config struct {
	NumServers          int
	MinTools            int
	MaxTools            int
	MinMutationInterval time.Duration
	MaxMutationInterval time.Duration
}

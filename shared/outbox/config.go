package outbox

import "time"

type Config struct {
	Period    time.Duration
	BatchSize int
}

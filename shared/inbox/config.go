package inbox

import "time"

type Config struct {
	Period    time.Duration
	BatchSize int
}

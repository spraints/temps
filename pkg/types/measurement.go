package types

import (
	"time"
)

type Measurement struct {
	ID          string
	Name        string
	Temperature Temperature
	MeasuredAt  time.Time
}

package event

import "time"

type Command struct {
	Time              time.Duration
	EventID, PlayerID uint8
	Arg               string
}

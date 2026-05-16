package dungeon

import "time"

type Config struct {
	Floors   uint8
	Monsters uint8
	OpenAt   time.Time
	Duration time.Duration
}

package game

import "time"

type Config struct {
	Floors   uint8  `json:"Floors"`
	Monsters uint8  `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration uint8  `json:"Duration"`
}

var Cfg Config

const timeLayout = "15:04:05"

func OpenDuration() (time.Duration, error) {
	parsed, err := time.Parse(timeLayout, Cfg.OpenAt)
	if err != nil {
		return 0, err
	}
	return time.Duration(parsed.Hour())*time.Hour +
		time.Duration(parsed.Minute())*time.Minute +
		time.Duration(parsed.Second())*time.Second, nil
}

func CloseDuration() (time.Duration, error) {
	openAt, err := OpenDuration()
	if err != nil {
		return 0, err
	}
	return openAt + time.Duration(Cfg.Duration)*time.Hour, nil
}

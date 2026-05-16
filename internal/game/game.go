package game

type Config struct {
	Floors   uint8  `json:"Floors"`
	Monsters uint8  `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration uint8  `json:"Duration"`
}

var Cfg Config

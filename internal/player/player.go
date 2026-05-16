package player

type PlayerStatus uint8

const (
	SUCCESS PlayerStatus = iota
	FAIL
	DISQUAL
)
const DefaultHealth uint8 = 20

type Player struct {
	ID, Health, Level uint8
	IsDisqualified    bool
}

var Players = make(map[uint8]Player)

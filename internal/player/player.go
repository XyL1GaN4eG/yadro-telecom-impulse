package player

type Status uint8

const (
	SUCCESS Status = iota
	FAIL
	DISQUAL
)
const DefaultHealth uint8 = 20

type Player struct {
	ID, Health, Level uint8
	IsDisqualified    bool
	Dungeon           DungeonRun
}

type DungeonRun struct {
	Floors     []Floor
	BossKilled bool
}

type Floor struct {
	MonstersLeft uint8
	Cleared      bool
}

var Players = make(map[uint8]Player)

package player

import "time"

const (
	MaxHealth     uint8 = 100
	DefaultHealth       = MaxHealth
	StartFloor    uint8 = 0
)

type Status uint8

const (
	StatusSuccess Status = iota
	StatusFail
	StatusDisqual
)

type Player struct {
	ID, Health, Floor        uint8
	Status                   Status
	EnteredDungeon, Finished bool
	EnteredAt, FinishedAt    time.Duration
	Dungeon                  DungeonRun
}

type DungeonRun struct {
	Floors           []Floor
	BossFloorEntered bool
	BossStartedAt    time.Duration
	BossKillTime     time.Duration
}

type Floor struct {
	MonstersLeft    uint8
	EnteredAt       time.Duration
	TimeSpent       time.Duration
	IsBoss, Cleared bool
	Entered         bool
}

func New(id uint8, floorsCount, monstersPerFloor uint8) Player {
	return Player{
		ID:     id,
		Health: DefaultHealth,
		Floor:  StartFloor,
		Status: StatusFail,
		Dungeon: NewDungeonRun(
			floorsCount,
			monstersPerFloor,
		),
	}
}

func NewDungeonRun(floorsCount, monstersPerFloor uint8) DungeonRun {
	floors := make([]Floor, int(floorsCount))

	for i := range floors {
		isBoss := i == len(floors)-1

		floors[i] = Floor{
			IsBoss:  isBoss,
			Cleared: false,
		}

		if !isBoss {
			floors[i].MonstersLeft = monstersPerFloor
			floors[i].Cleared = monstersPerFloor == 0
		}
	}

	return DungeonRun{
		Floors:           floors,
		BossFloorEntered: false,
	}
}

func (p Player) TotalTime() time.Duration {
	if !p.EnteredDungeon || p.FinishedAt < p.EnteredAt {
		return 0
	}
	return p.FinishedAt - p.EnteredAt
}

func (p Player) AverageFloorClearTime() time.Duration {
	var total time.Duration
	var cleared uint8

	for _, floor := range p.Dungeon.Floors {
		if floor.IsBoss || !floor.Cleared {
			continue
		}
		total += floor.TimeSpent
		cleared++
	}

	if cleared == 0 {
		return 0
	}
	return total / time.Duration(cleared)
}

var Players = make(map[uint8]Player)

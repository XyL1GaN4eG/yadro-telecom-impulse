package player

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
	Dungeon                  DungeonRun
}

type DungeonRun struct {
	Floors []Floor
}

type Floor struct {
	MonstersLeft    uint8
	IsBoss, Cleared bool
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
		}
	}

	return DungeonRun{
		Floors: floors,
	}
}

var Players = make(map[uint8]Player)

package handler

import (
	"errors"
	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/parser"
	"impulse/internal/player"
	"time"
)

var ErrNoOutput = errors.New("event produces no output")

func HandleCommand(cmd event.Command) (p player.Player, err error) {
	if cmd.EventID != 1 {
		if _, ok := player.Players[cmd.PlayerID]; !ok {
			return disqualifyPlayer(cmd), nil
		}
	}

	switch cmd.EventID {
	case 1:
		p, err = registerPlayer(cmd)
	case 2:
		p, err = enterToDungeon(cmd)
	case 3:
		p, err = killEnemy(cmd)
	case 4:
		p, err = moveToNextFloor(cmd)
	case 5:
		p, err = moveToPrevFloor(cmd)
	case 6:
		p, err = enterBossLevel(cmd)
	case 7:
		p, err = killBoss(cmd)
	case 8:
		p, err = leaveDungeon(cmd)
	case 9:
		p, err = cannotContinue(cmd)
	case 10:
		p, err = heal(cmd)
	case 11:
		p, err = damage(cmd)
	case 33:
		p, err = killPlayer(cmd)
	default:
		err = errors.New("unknown event")
	}
	return
}

func registerPlayer(cmd event.Command) (player.Player, error) {
	if _, ok := player.Players[cmd.PlayerID]; ok {
		return player.Player{}, errors.New("player already register")
	}

	p := player.New(
		cmd.PlayerID,
		game.Cfg.Floors,
		game.Cfg.Monsters,
	)

	player.Players[cmd.PlayerID] = p
	return p, nil
}

func enterToDungeon(cmd event.Command) (p player.Player, err error) {
	p, ok := player.Players[cmd.PlayerID]
	if !ok {
		return disqualifyPlayer(cmd), nil
	}
	if p.Health <= 0 || p.Status == player.StatusDisqual || p.Finished {
		return p, ErrNoOutput
	}

	if p.EnteredDungeon {
		return player.Player{}, errors.New("player already entered dungeon")
	}

	p.Floor = 0
	p.EnteredDungeon = true
	p.EnteredAt = cmd.Time
	enterFloor(&p, cmd.Time)
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func killEnemy(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return player.Player{}, errors.New("incorrect floor")
	}

	currentFloor := &p.Dungeon.Floors[p.Floor]
	if currentFloor.IsBoss {
		return p, errors.New("boss floor doesn't have monsters")
	}
	if currentFloor.MonstersLeft == 0 {
		err = errors.New("you don't have any monsters")
		return p, err
	}

	currentFloor.MonstersLeft--
	if currentFloor.MonstersLeft == 0 {
		currentFloor.Cleared = true
		completeFloor(currentFloor, cmd.Time)
	}
	player.Players[p.ID] = p
	return
}

func moveToNextFloor(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}

	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return player.Player{}, errors.New("incorrect floor")
	}

	if !p.Dungeon.Floors[p.Floor].Cleared {
		err = errors.New("floor isn't cleared")
		return player.Player{}, err
	}

	nextFloor := int(p.Floor) + 1
	if nextFloor >= len(p.Dungeon.Floors) {
		return p, errors.New("person already on a last level")
	}

	leaveFloor(&p, cmd.Time)
	p.Floor = uint8(nextFloor)
	enterFloor(&p, cmd.Time)
	if p.Dungeon.Floors[p.Floor].IsBoss && !p.Dungeon.BossFloorEntered {
		startBoss(&p, cmd.Time)
	}
	player.Players[p.ID] = p
	return p, nil
}

func moveToPrevFloor(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if p.Floor > 0 {
		leaveFloor(&p, cmd.Time)
		p.Floor--
		enterFloor(&p, cmd.Time)
		player.Players[p.ID] = p
		return p, nil
	}

	return p, errors.New("person already on a first level")
}

func enterBossLevel(cmd event.Command) (player.Player, error) {
	p, err := findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return player.Player{}, errors.New("incorrect floor")
	}
	if !p.Dungeon.Floors[p.Floor].IsBoss {
		return player.Player{}, errors.New("floor isn't boss")
	}
	if p.Dungeon.BossFloorEntered {
		return p, ErrNoOutput
	}
	startBoss(&p, cmd.Time)
	player.Players[p.ID] = p

	return p, nil
}

func killBoss(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return player.Player{}, errors.New("incorrect floor")
	}
	if !p.Dungeon.Floors[p.Floor].IsBoss {
		return p, errors.New("floor isn't boss")
	}
	if !p.Dungeon.BossFloorEntered {
		return p, errors.New("boss floor wasn't entered")
	}

	p.Status = player.StatusSuccess
	p.Dungeon.Floors[p.Floor].Cleared = true
	if cmd.Time >= p.Dungeon.BossStartedAt {
		p.Dungeon.BossKillTime = cmd.Time - p.Dungeon.BossStartedAt
	}
	player.Players[p.ID] = p
	return p, nil
}

func leaveDungeon(cmd event.Command) (player.Player, error) {
	p, err := findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	finishPlayer(&p, cmd.Time)
	player.Players[p.ID] = p
	return p, nil
}

func cannotContinue(cmd event.Command) (player.Player, error) {
	p, ok := player.Players[cmd.PlayerID]
	if !ok {
		return disqualifyPlayer(cmd), nil
	}
	if p.Finished {
		return p, ErrNoOutput
	}
	p.Status = player.StatusDisqual
	finishPlayer(&p, cmd.Time)
	player.Players[p.ID] = p
	return p, nil
}

func heal(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	health, err := parser.StrToUint8(cmd.Arg)
	if err != nil {
		err = errors.New("failed to parse health")
		return
	}
	newHealth := int(p.Health) + int(health)
	if newHealth > int(player.MaxHealth) {
		newHealth = int(player.MaxHealth)
	}
	p.Health = uint8(newHealth)
	player.Players[cmd.PlayerID] = p
	return
}

func damage(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	damage, err := parser.StrToUint8(cmd.Arg)
	if err != nil {
		err = errors.New("failed to parse damage")
		return
	}
	if p.Health <= damage {
		p.Health = 0
		p.Status = player.StatusFail
		finishPlayer(&p, cmd.Time)
	} else {
		p.Health -= damage
	}
	player.Players[cmd.PlayerID] = p
	return
}

func killPlayer(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}
	p.Health = 0
	p.Status = player.StatusFail
	finishPlayer(&p, cmd.Time)
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func findLivePlayer(playerID uint8) (player.Player, error) {
	p, ok := player.Players[playerID]
	if !ok || p.Health == 0 || p.Status == player.StatusDisqual || p.Finished {
		return p, ErrNoOutput
	}
	if !p.EnteredDungeon {
		return p, errors.New("person not registered yet or not alive or not entered dungeon")
	}
	return p, nil
}

func disqualifyPlayer(cmd event.Command) player.Player {
	p := player.New(
		cmd.PlayerID,
		game.Cfg.Floors,
		game.Cfg.Monsters,
	)
	p.Status = player.StatusDisqual
	finishPlayer(&p, cmd.Time)
	player.Players[cmd.PlayerID] = p
	return p
}

func CloseActivePlayers(at time.Duration) {
	for id, p := range player.Players {
		if p.Finished {
			continue
		}
		finishPlayer(&p, at)
		player.Players[id] = p
	}
}

func finishPlayer(p *player.Player, at time.Duration) {
	p.Finished = true
	p.FinishedAt = at
}

func enterFloor(p *player.Player, at time.Duration) {
	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return
	}

	floor := &p.Dungeon.Floors[p.Floor]
	if floor.IsBoss || floor.Cleared || floor.Entered {
		return
	}
	floor.Entered = true
	floor.EnteredAt = at
}

func leaveFloor(p *player.Player, at time.Duration) {
	if int(p.Floor) >= len(p.Dungeon.Floors) {
		return
	}

	floor := &p.Dungeon.Floors[p.Floor]
	if floor.IsBoss || floor.Cleared || !floor.Entered || at < floor.EnteredAt {
		return
	}
	floor.TimeSpent += at - floor.EnteredAt
	floor.Entered = false
}

func completeFloor(floor *player.Floor, at time.Duration) {
	if !floor.Entered || at < floor.EnteredAt {
		return
	}
	floor.TimeSpent += at - floor.EnteredAt
	floor.Entered = false
}

func startBoss(p *player.Player, at time.Duration) {
	p.Dungeon.BossFloorEntered = true
	p.Dungeon.BossStartedAt = at
}

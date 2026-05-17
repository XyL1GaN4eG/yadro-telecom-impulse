package handler

import (
	"errors"
	"impulse/internal/event"
	"impulse/internal/game"
	"impulse/internal/parser"
	"impulse/internal/player"
	"log"
	"strconv"
)

func HandleCommand(cmd event.Command) (p player.Player, err error) {
	switch cmd.EventID {
	case 1:
		p, err = registerPlayer(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v зарегистрирован\n", p.ID)
		}
	case 2:
		p, err = enterToDungeon(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v вошел в подземелье", p.ID)
		}
	case 3:
		p, err = killEnemy(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v убил монстра", p.ID)
		}
	case 4:
		p, err = moveToNextFloor(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v перешел на следующий этаж", p.ID)
		}
	case 5:
		p, err = moveToPrevFloor(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v перешел на предыдущий этаж", p.ID)
		}
	case 6:
		p, err = enterBossLevel(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v вошел на этаж босса", p.ID)
		}

	case 7:
		p, err := killBoss(cmd)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Игрок %v убил босса", p.ID)
		}
	case 8:
		p, err = leaveDungeon(cmd)
	case 9:
		p, err = cannotContinue(cmd)
	case 10:
		_, err = heal(cmd)
	case 11:
		_, err = damage(cmd)
	case 33:
		_, err = killPlayer(cmd)
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
		return disqualifyPlayer(cmd.PlayerID), nil
	}
	if p.Health <= 0 {
		return player.Player{}, errors.New("player is dead")
	}

	if p.EnteredDungeon {
		return player.Player{}, errors.New("player already entered dungeon")
	}

	p.Floor = 0
	p.EnteredDungeon = true
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

	p.Floor = uint8(nextFloor)
	player.Players[p.ID] = p
	return p, nil
}

func moveToPrevFloor(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if p.Floor > 0 {
		p.Floor--
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
	p.Dungeon.BossFloorEntered = true
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
	player.Players[p.ID] = p
	return p, nil
}

func leaveDungeon(cmd event.Command) (player.Player, error) {
	p, err := findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	p.Finished = true
	player.Players[p.ID] = p
	return p, nil
}

func cannotContinue(cmd event.Command) (player.Player, error) {
	p, ok := player.Players[cmd.PlayerID]
	if !ok {
		return disqualifyPlayer(cmd.PlayerID), nil
	}
	p.Status = player.StatusDisqual
	p.Finished = true
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
	} else {
		p.Health -= damage
	}
	player.Players[cmd.PlayerID] = p
	return
}

func killPlayer(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, errors.New("player=" + strconv.Itoa(int(cmd.PlayerID)) + "wasn't found")
	}
	p.Health = 0
	p.Status = player.StatusFail
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func findLivePlayer(playerID uint8) (player.Player, error) {
	p, ok := player.Players[playerID]
	if !ok || p.Health == 0 || p.Status == player.StatusDisqual || !p.EnteredDungeon {
		return p, errors.New("person not registered yet or not alive or not entered dungeon")
	}
	return p, nil
}

func disqualifyPlayer(playerID uint8) player.Player {
	p := player.New(
		playerID,
		game.Cfg.Floors,
		game.Cfg.Monsters,
	)
	p.Status = player.StatusDisqual
	p.Finished = true
	player.Players[playerID] = p
	return p
}

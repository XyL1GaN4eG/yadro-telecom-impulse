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
		}
		log.Printf("Игрок %v зарегистрирован\n", p.ID)
	case 2:
		p, err = enterToDungeon(cmd)
		if err != nil {
		} else {
			log.Printf("Игрок %v вошел в подземелье", p.ID)
		}
	case 3:
		p, err = killEnemy(cmd)
		if err != nil {

		} else {
			log.Printf("Игрок %v убил монстра", p.ID)
		}
	case 4:
		p, err = moveToNextFloor(cmd)
		if err != nil {
		} else {
			log.Printf("Игрок %v перешел на следующий этаж", p.ID)
		}
	case 5:
		p, err = moveToPrevFloor(cmd)
		if err != nil {
		} else {
			log.Printf("Игрок %v перешел на предыдущий этаж", p.ID)
		}
	case 6:
		log.Printf("Игрок %v вошел на этаж босса", p.ID)
	case 7:
		log.Printf("Игрок %v убил босса", p.ID)
	case 8:
		log.Printf("Игрок %v покинул подземелье", p.ID)
	case 9:
		log.Printf("Игрок %v не может продолжать из-за [%s]", p.ID, cmd.Arg)
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
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	p.Floor = 1
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func killEnemy(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	return
}

func moveToNextFloor(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if p.Floor < uint8(len(p.Dungeon.Floors)) {
		p.Floor++
		player.Players[p.ID] = p
		return p, nil
	}

	return p, errors.New("person already on a last level")
}

func moveToPrevFloor(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return player.Player{}, err
	}
	if !(p.Floor < 1) {
		p.Floor--
		player.Players[p.ID] = p
		return p, nil
	}

	return p, errors.New("person already on a first level")
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
	p.Health += health
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
	} else {
		p.Health -= damage
	}
	player.Players[cmd.PlayerID] = p
	return
}

func killPlayer(cmd event.Command) (p player.Player, err error) {
	p, ok := player.Players[cmd.PlayerID]
	if !ok {
		return player.Player{}, errors.New("player=" + strconv.Itoa(int(cmd.PlayerID)) + "wasn't found")
	}
	p.Health = 0
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func findLivePlayer(playerID uint8) (player.Player, error) {
	p, ok := player.Players[playerID]
	if !ok || p.Health <= 1 {
		return p, errors.New("person not registered yet or already dead")
	}
	return p, nil
}

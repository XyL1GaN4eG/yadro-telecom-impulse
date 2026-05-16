package handler

import (
	"errors"
	"impulse/internal/event"
	"impulse/internal/parser"
	"impulse/internal/player"
	"strconv"
)

func HandleCommand(cmd event.Command, cfg app.GameConfig) (err error) {
	switch cmd.EventID {
	case 1:
		_, err = registerPlayer(cmd)
	case 2:
		_, err = enterToDungeon(cmd)
	case 3:
		_, err = killEnemy(cmd, cfg)
	case 4:
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

func registerPlayer(cmd event.Command) (p player.Player, err error) {
	_, ok := player.Players[cmd.PlayerID]
	if !ok {
		p = player.Player{
			ID:             cmd.PlayerID,
			Health:         player.DefaultHealth,
			IsDisqualified: false,
		}
		player.Players[cmd.PlayerID] = p
		return
	}
	err = errors.New("player already register")
	return player.Player{}, err
}

func enterToDungeon(cmd event.Command) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	p.Level = 1
	player.Players[cmd.PlayerID] = p
	return p, nil
}

func killEnemy(cmd event.Command, cfg app.GameConfig) (p player.Player, err error) {
	p, err = findLivePlayer(cmd.PlayerID)
	if err != nil {
		return p, err
	}

	return
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

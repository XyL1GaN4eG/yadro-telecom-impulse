package app

import (
	"bufio"
	"errors"
	"fmt"
	"impulse/internal/io"
	"os"
	"strconv"
	"strings"
	"time"
)

type playerStatus uint8

const (
	SUCCESS playerStatus = iota
	FAIL
	DISQUAL
)
const defaultHealth uint8 = 20

type GameConfig struct {
	Floors   uint8  `json:"Floors"`
	Monsters uint8  `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration uint8  `json:"Duration"`
}

func Run() {
	for {
		now := time.Now()
		timeStr := now.Format("2006-01-02 15:04:05")

		fmt.Printf("[%v]", timeStr)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			cmd, err := split(line)
			if err != nil {
			}
			err = handleCommand(cmd)
			if err != nil {
				_ = io.PrintError(err)
			}
		}

	}
}

type command struct {
	eventID, playerID uint8
	arg               string
}

type Player struct {
	id, health, level uint8
	isDisqualified    bool
}

var players map[uint8]Player

func split(line string) (cmd command, err error) {
	cmd = command{
		eventID:  0,
		playerID: 0,
		arg:      "",
	}
	parts := strings.Fields(line)
	l := len(parts)

	if l > 2 || l == 0 {
		err = errors.New("expected one or two args")
		goto end
	}

	cmd.eventID, err = strToUint8(parts[0])
	if err != nil || !(cmd.eventID <= 11 || (cmd.eventID >= 31 && cmd.eventID <= 33)) {
		err = errors.New("failed to parse eventID")
		goto end
	}

	cmd.playerID, err = strToUint8(parts[1])
	if err != nil {
		err = errors.New("failed to parse userID")
	}

	if cmd.eventID >= 9 && cmd.eventID <= 11 && l == 2 {
		cmd.playerID, err = strToUint8(parts[1])
		if err != nil {
			cmd.eventID = 0
			err = errors.New("incorrect arg for command=" + string(cmd.eventID))
		}
		goto end
	}
	return command{}, errors.New("failed to parse line=" + line)
end:
	return cmd, nil
}

func handleCommand(cmd command) (err error) {
	switch cmd.eventID {
	case 1:
		_, err = registerPlayer(cmd)
	case 2:
		_, err = enterToDungeon(cmd)
	case 3:
	case 4:
	default:
		err = errors.New("unknown event")
	}
	return
}

func isPlayerExists(id uint8) bool {
	_, ok := players[id]
	return ok
}

func registerPlayer(cmd command) (player Player, err error) {
	if !isPlayerExists(cmd.playerID) {
		player = Player{
			id:             cmd.playerID,
			health:         defaultHealth,
			isDisqualified: false,
		}
		players[cmd.playerID] = player
		return
	}
	err = errors.New("player already register")
	return Player{}, err
}

func enterToDungeon(cmd command) (player Player, err error) {
	player, ok := players[cmd.playerID]
	if !ok {
		return Player{}, errors.New("player don't register yet")
	}
	player.level = 1
	return player, nil
}

func strToUint8(s string) (arg uint8, err error) {
	parseUint, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(parseUint), nil
}

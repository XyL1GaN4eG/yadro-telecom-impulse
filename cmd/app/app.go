package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"impulse/internal/event"
	outfmt "impulse/internal/format"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/parser"
	"impulse/internal/player"
	"log"
	"os"
)

func Run() {
	args := os.Args[1:]
	var path string
	if len(args) != 0 {
		path = args[0]
	} else {
		path = "docs/config.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}
	if err := json.Unmarshal(data, &game.Cfg); err != nil {
		panic("Error parsing JSON:" + string(err.Error()))
	}
	closeAt, err := game.CloseDuration()
	if err != nil {
		log.Panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var lastEventTime int64 = -1
	closed := false
	for scanner.Scan() {
		line := scanner.Text()

		cmd, err := parser.Split(line)
		if err != nil {
			log.Println(err)
			continue
		}
		if lastEventTime > int64(cmd.Time) {
			log.Println("events must be sequential in time")
			continue
		}
		lastEventTime = int64(cmd.Time)

		if !closed && cmd.Time >= closeAt {
			handler.CloseActivePlayers(closeAt)
			closed = true
		}
		if closed {
			continue
		}

		wasBossFloorEntered := false
		if p, ok := player.Players[cmd.PlayerID]; ok {
			wasBossFloorEntered = p.Dungeon.BossFloorEntered
		}
		p, err := handler.HandleCommand(cmd)
		if err != nil {
			if errors.Is(err, handler.ErrNoOutput) {
				continue
			}
			if cmd.EventID >= 4 && cmd.EventID <= 7 {
				fmt.Println(outfmt.ImpossibleMove(cmd))
			}
			continue
		}

		if p.Status == player.StatusDisqual {
			fmt.Println(outfmt.Disqualified(cmd))
			continue
		}
		fmt.Println(outfmt.Command(cmd))
		if cmd.EventID == 4 && !wasBossFloorEntered && p.Dungeon.BossFloorEntered {
			fmt.Println(outfmt.Command(event.Command{
				Time:     cmd.Time,
				PlayerID: cmd.PlayerID,
				EventID:  6,
			}))
		}
		if cmd.EventID == 11 && p.Health == 0 {
			fmt.Println(outfmt.Dead(cmd))
		}
	}
	handler.CloseActivePlayers(closeAt)
	if len(player.Players) > 0 {
		fmt.Print(outfmt.FinalReport(player.Players))
	}
}

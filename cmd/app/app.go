package app

import (
	"bufio"
	"fmt"
	outfmt "impulse/internal/format"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/parser"
	"impulse/internal/player"
	"impulse/internal/replay"
	"impulse/internal/tui"
	"log"
	"os"
)

func Run() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "tui" {
		if err := tui.Run(args[1:]); err != nil {
			log.Panic(err)
		}
		return
	}

	var path string
	if len(args) != 0 {
		path = args[0]
	} else {
		path = "docs/config.json"
	}
	if err := replay.LoadConfig(path); err != nil {
		log.Panic(err)
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

		for _, line := range replay.ProcessCommand(cmd) {
			fmt.Println(line)
		}
	}
	handler.CloseActivePlayers(closeAt)
	if len(player.Players) > 0 {
		fmt.Print(outfmt.FinalReport(player.Players))
	}
}

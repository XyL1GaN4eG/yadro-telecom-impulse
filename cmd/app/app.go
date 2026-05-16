package app

import (
	"bufio"
	"fmt"
	"impulse/internal/dungeon"
	"impulse/internal/handler"
	"impulse/internal/io"
	"impulse/internal/parser"
	"os"
	"time"
)

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
		cfg := parseGameConfig()
		for scanner.Scan() {
			line := scanner.Text()
			cmd, err := parser.Split(line)
			if err != nil {
			}
			err = handler.HandleCommand(cmd, cfg)
			if err != nil {
				_ = io.PrintError(err)
			}
		}

	}
}

func parseGameConfig() dungeon.Dungeon {
	return dungeon.Dungeon{
		Floor: nil,
	}
}

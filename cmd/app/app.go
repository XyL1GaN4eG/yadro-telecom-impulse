package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/io"
	"impulse/internal/parser"
	"os"
	"time"
)

func Run() {
	data, err := os.ReadFile("config.json")
	if err != nil {
	}
	if err := json.Unmarshal(data, &game.Cfg); err != nil {
		panic("Error parsing JSON:" + string(err.Error()))
		return
	}

	for {
		now := time.Now()
		timeStr := now.Format("2006-01-02 15:04:05")

		fmt.Printf("[%v]", timeStr)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			cmd, err := parser.Split(line)
			if err != nil {
			}
			err = handler.HandleCommand(cmd)
			if err != nil {
				_ = io.PrintError(err)
			}
		}

	}
}

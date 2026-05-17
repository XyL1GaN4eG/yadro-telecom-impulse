package app

import (
	"bufio"
	"encoding/json"
	"fmt"
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
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		cmd, err := parser.Split(line)
		if err != nil {
			log.Println(err)
			continue
		}

		p, err := handler.HandleCommand(cmd)
		if err != nil {
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
		if cmd.EventID == 11 && p.Health == 0 {
			fmt.Println(outfmt.Dead(cmd))
		}
	}
}

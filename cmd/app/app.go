package app

import (
	"bufio"
	"fmt"
	"impulse/internal/parser"
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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		cmd, err := parser.Split(line)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, line := range replay.ProcessCommand(cmd) {
			fmt.Println(line)
		}
	}
}

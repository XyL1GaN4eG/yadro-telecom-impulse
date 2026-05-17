package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/io"
	"impulse/internal/parser"
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

	fmt.Printf("%v\n", game.Cfg)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		log.Println("]]]")

		line := scanner.Text()

		cmd, err := parser.Split(line)
		if err != nil {
			log.Println(err)
		}
		_, _ = handler.HandleCommand(cmd)
		if err != nil {
			_ = io.PrintError(err)
		}
	}
}

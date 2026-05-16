package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Run() {
	for {
		now := time.Now()
		timeStr := now.Format("2006-01-02 15:04:05")

		fmt.Printf("[%v]", timeStr)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			//line := scanner.Text()
			//cmd, arg, err := split(line)
			//switch cmd {
			//1:
		}
	}

}

//}

type cmd struct {
	eventID  uint8
	playerID uint8
	arg      string
}

func split(line string) (cmd, arg uint8, err error) {
	cmd = 0
	arg = 0
	parts := strings.Fields(line)
	l := len(parts)
	if l > 2 || l == 0 {
		err = errors.New("expected one or two args")
		goto end
	}

	cmd, err = strToUint8(parts[0])
	if err != nil {
		err = errors.New("failed to parse command")
		goto end
	}

	if !(cmd <= 11 || (cmd >= 31 && cmd <= 33)) {
		err = errors.New("incorrect number of command")
		cmd = 0
		goto end
	}

	if cmd >= 9 && cmd <= 11 && l == 2 {
		arg, err = strToUint8(parts[1])
		if err != nil {
			cmd = 0
			err = errors.New("incorrect arg for cmd=" + string(cmd))
		}
		goto end
	}
	return 0, 0, errors.New("failed to parse line=" + line)
end:
	return cmd, arg, nil
}

func strToUint8(s string) (arg uint8, err error) {
	parseUint, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(parseUint), nil
}

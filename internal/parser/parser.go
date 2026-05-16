package parser

import (
	"errors"
	"impulse/internal/event"
	"strconv"
	"strings"
)

func Split(line string) (cmd event.Command, err error) {
	/*
	   todo: добавить парсинг времени
	*/
	cmd = event.Command{
		EventID:  0,
		PlayerID: 0,
		Arg:      "",
	}
	parts := strings.Fields(line)
	l := len(parts)

	if l > 2 || l == 0 { // todo: убрать l>2, тк при reason может передаваться несколько слов
		err = errors.New("expected one or two args")
		goto end
	}

	cmd.EventID, err = StrToUint8(parts[0])
	if err != nil || !(cmd.EventID <= 11 || (cmd.EventID >= 31 && cmd.EventID <= 33)) {
		err = errors.New("failed to parse EventID")
		goto end
	}

	cmd.PlayerID, err = StrToUint8(parts[1])
	if err != nil {
		err = errors.New("failed to parse userID")
	}

	if cmd.EventID >= 9 && cmd.EventID <= 11 && l == 2 {
		cmd.PlayerID, err = StrToUint8(parts[1])
		if err != nil {
			cmd.EventID = 0
			err = errors.New("incorrect arg for Command=" + strconv.Itoa(int(cmd.EventID)))
		}
		goto end
	}
	return event.Command{}, errors.New("failed to parse line=" + line)
end:
	return cmd, nil
}

func StrToUint8(s string) (arg uint8, err error) {
	parseUint, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(parseUint), nil
}

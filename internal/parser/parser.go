package parser

import (
	"errors"
	"impulse/internal/event"
	"strconv"
	"strings"
	"time"
)

const timeLayout = "15:04:05"

func Split(line string) (cmd event.Command, err error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return event.Command{}, errors.New("expected time, player id and event id")
	}

	if len(parts[0]) != len("[00:00:00]") || parts[0][0] != '[' || parts[0][len(parts[0])-1] != ']' {
		return event.Command{}, errors.New("failed to parse time")
	}

	parsedTime, err := time.Parse(timeLayout, strings.Trim(parts[0], "[]"))
	if err != nil {
		return event.Command{}, errors.New("failed to parse time")
	}

	playerID, err := StrToUint8(parts[1])
	if err != nil {
		return event.Command{}, errors.New("failed to parse userID")
	}

	eventID, err := StrToUint8(parts[2])
	if err != nil || eventID > 11 {
		return event.Command{}, errors.New("failed to parse EventID")
	}

	return event.Command{
		Time:     DurationSinceMidnight(parsedTime),
		EventID:  eventID,
		PlayerID: playerID,
		Arg:      strings.Join(parts[3:], " "),
	}, nil
}

func StrToUint8(s string) (arg uint8, err error) {
	parseUint, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(parseUint), nil
}

func DurationSinceMidnight(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second
}

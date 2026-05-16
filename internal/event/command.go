package event

type Command struct {
	EventID, PlayerID uint8
	Arg               string
}

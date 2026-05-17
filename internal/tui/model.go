package tui

import (
	"bufio"
	"fmt"
	"impulse/internal/event"
	"impulse/internal/format"
	"impulse/internal/game"
	"impulse/internal/handler"
	"impulse/internal/parser"
	"impulse/internal/player"
	"impulse/internal/replay"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const tickEvery = 500 * time.Millisecond

type Entry struct {
	Line string
	Cmd  event.Command
	Err  error
}

type Model struct {
	Entries    []Entry
	Index      int
	Log        []string
	Playing    bool
	LastErr    string
	ShowHelp   bool
	Width      int
	Height     int
	CloseAt    time.Duration
	Closed     bool
	Reported   bool
	lastTime   int64
	activeID   uint8
	selectedID uint8
	keys       keyMap
	help       help.Model
	zoneID     string
}

type tickMsg time.Time

func Run(args []string) error {
	configPath := "docs/config.json"
	eventsPath := "docs/events"
	if len(args) > 0 {
		configPath = args[0]
	}
	if len(args) > 1 {
		eventsPath = args[1]
	}

	if err := replay.LoadConfig(configPath); err != nil {
		return err
	}
	entries, err := LoadEvents(eventsPath)
	if err != nil {
		return err
	}
	replay.ResetPlayers()
	zone.NewGlobal()
	defer zone.Close()

	_, err = tea.NewProgram(
		NewModel(entries),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	).Run()
	return err
}

func LoadEvents(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		cmd, err := parser.Split(line)
		entries = append(entries, Entry{Line: line, Cmd: cmd, Err: err})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func NewModel(entries []Entry) Model {
	zone.NewGlobal()
	helpModel := help.New()
	helpModel.ShowAll = true
	closeAt, _ := game.CloseDuration()
	return Model{
		Entries:  entries,
		Width:    100,
		Height:   32,
		CloseAt:  closeAt,
		lastTime: -1,
		keys:     defaultKeys,
		help:     helpModel,
		zoneID:   zone.NewPrefix(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type keyMap struct {
	Forward key.Binding
	Back    key.Binding
	Play    key.Binding
	Reset   key.Binding
	End     key.Binding
	Help    key.Binding
	Quit    key.Binding
	Mouse   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Forward, k.Back, k.Play, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Forward, k.Back, k.Play, k.Reset, k.End},
		{k.Help, k.Quit, k.Mouse},
	}
}

var defaultKeys = keyMap{
	Forward: key.NewBinding(
		key.WithKeys("l", "k", "right"),
		key.WithHelp("l/k", "step forward"),
	),
	Back: key.NewBinding(
		key.WithKeys("h", "j", "left"),
		key.WithHelp("h/j", "step back"),
	),
	Play: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "play/pause"),
	),
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset"),
	),
	End: key.NewBinding(
		key.WithKeys("g", "G"),
		key.WithHelp("g", "to end"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Mouse: key.NewBinding(
		key.WithKeys("click"),
		key.WithHelp("click", "select player"),
	),
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		case "esc":
			m.ShowHelp = false
			return m, nil
		}
		if m.ShowHelp {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Play):
			m.Playing = !m.Playing
			if m.Playing {
				return m, tick()
			}
		case key.Matches(msg, m.keys.Forward):
			m.Step()
		case key.Matches(msg, m.keys.Back):
			m.StepBack()
		case key.Matches(msg, m.keys.Reset):
			m.Reset()
		case key.Matches(msg, m.keys.End):
			m.PlayToEnd()
		}
	case tea.MouseMsg:
		if m.ShowHelp || msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		for _, id := range playerIDs() {
			z := zone.Get(m.playerZone(id))
			if z != nil && z.InBounds(msg) {
				m.selectedID = id
				m.activeID = id
				return m, nil
			}
		}
	case tickMsg:
		if m.Playing {
			if !m.Step() {
				m.Playing = false
				return m, nil
			}
			return m, tick()
		}
	}
	return m, nil
}

func tick() tea.Cmd {
	return tea.Tick(tickEvery, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Step() bool {
	if m.Index >= len(m.Entries) {
		m.Playing = false
		return false
	}

	entry := m.Entries[m.Index]
	m.LastErr = ""
	if entry.Err != nil {
		m.LastErr = entry.Err.Error()
		m.Log = append(m.Log, fmt.Sprintf("parse error at step %d: %v", m.Index+1, entry.Err))
		m.Index++
		return true
	}
	if m.lastTime > int64(entry.Cmd.Time) {
		m.LastErr = "events must be sequential in time"
		m.Log = append(m.Log, fmt.Sprintf("parse error at step %d: events must be sequential in time", m.Index+1))
		m.Index++
		return true
	}
	m.lastTime = int64(entry.Cmd.Time)
	if !m.Closed && m.CloseAt > 0 && entry.Cmd.Time >= m.CloseAt {
		handler.CloseActivePlayers(m.CloseAt)
		m.Closed = true
		m.appendFinalReport()
		m.Index = len(m.Entries)
		m.Playing = false
		return true
	}
	if m.Closed {
		m.Index++
		return true
	}

	m.activeID = entry.Cmd.PlayerID
	if m.selectedID == 0 {
		m.selectedID = entry.Cmd.PlayerID
	}
	lines := replay.ProcessCommand(entry.Cmd)
	if len(lines) == 0 {
		m.LastErr = "event ignored"
	}
	m.Log = append(m.Log, lines...)
	m.Index++
	if m.Index >= len(m.Entries) {
		m.closeAtEnd()
	}
	return true
}

func (m *Model) StepBack() {
	if m.Index == 0 {
		return
	}
	target := m.Index - 1
	m.replayTo(target)
}

func (m *Model) Reset() {
	m.replayTo(0)
}

func (m *Model) PlayToEnd() {
	m.replayTo(len(m.Entries))
	m.Playing = false
}

func (m *Model) replayTo(target int) {
	replay.ResetPlayers()
	m.Index = 0
	m.Log = nil
	m.LastErr = ""
	m.activeID = 0
	m.Closed = false
	m.Reported = false
	m.lastTime = -1
	m.Playing = false
	for m.Index < target {
		m.Step()
	}
}

func (m *Model) closeAtEnd() {
	if !m.Closed && m.CloseAt > 0 {
		handler.CloseActivePlayers(m.CloseAt)
		m.Closed = true
	}
	m.appendFinalReport()
}

func (m *Model) appendFinalReport() {
	if m.Reported || len(player.Players) == 0 {
		return
	}
	m.Log = append(m.Log, strings.Split(strings.TrimSuffix(format.FinalReport(player.Players), "\n"), "\n")...)
	m.Reported = true
}

func (m Model) View() string {
	width := max(m.Width, 80)
	height := max(m.Height, 24)
	if m.ShowHelp {
		return m.helpView(width, height)
	}
	panelHeight := max(height-9, 10)
	timelineWidth := clamp(width/4, 24, 36)
	playersWidth := clamp(width/3, 30, 44)
	centerWidth := max(width-timelineWidth-playersWidth-4, 24)
	logHeight := max(height-panelHeight-4, 5)

	top := headerStyle.Width(width).Render(m.header())
	timeline := panelStyle.Width(timelineWidth).Height(panelHeight).Render(m.timelineView(panelHeight))
	dungeon := panelStyle.Width(centerWidth).Height(panelHeight).Render(m.dungeonView(centerWidth, panelHeight))
	players := playersPanelStyle.Width(playersWidth).Height(panelHeight).Render(m.playersView(panelHeight))
	log := logStyle.Width(width).Height(logHeight).Render(m.logView(logHeight))
	return zone.Scan(lipgloss.JoinVertical(lipgloss.Left, top, lipgloss.JoinHorizontal(lipgloss.Top, timeline, dungeon, players), log))
}

func (m Model) header() string {
	mode := "paused"
	if m.Playing {
		mode = "playing"
	}
	errText := ""
	if m.LastErr != "" {
		errText = " | last: " + m.LastErr
	}
	return fmt.Sprintf(
		"Floors %d | Monsters %d | Open %s | Close %s | Step %d/%d | %s | ? help%s",
		game.Cfg.Floors,
		game.Cfg.Monsters,
		game.Cfg.OpenAt,
		format.Duration(m.CloseAt),
		m.Index,
		len(m.Entries),
		mode,
		errText,
	)
}

func (m Model) helpView(width, height int) string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Replay controls"),
		"",
		m.help.View(m.keys),
		"",
		mutedStyle.Render("Click a player row to inspect that player's dungeon path."),
	)
	help := helpStyle.Width(clamp(width-12, 36, 56)).Render(body)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, help)
}

func (m Model) timelineView(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Timeline"))
	b.WriteByte('\n')
	start := 0
	visible := max(height-2, 1)
	if m.Index >= visible {
		start = m.Index - visible + 1
	}
	end := min(len(m.Entries), start+visible)
	for i := start; i < end; i++ {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.Index {
			prefix = "> "
			style = currentStyle
		} else if i < m.Index {
			prefix = "x "
			style = doneStyle
		}
		line := fmt.Sprintf("%s%02d %s", prefix, i+1, m.eventLabel(m.Entries[i]))
		b.WriteString(style.Render(line))
		if i < end-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) eventLabel(entry Entry) string {
	if entry.Err != nil {
		return "parse error"
	}
	return fmt.Sprintf("%s P%d E%d", format.Time(entry.Cmd.Time), entry.Cmd.PlayerID, entry.Cmd.EventID)
}

func (m Model) dungeonView(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Dungeon"))
	b.WriteByte('\n')
	contentWidth := clamp(width-6, 22, 44)
	selected, hasSelected := m.selectedPlayer()
	if hasSelected {
		fmt.Fprintf(&b, "Inspect P%d | HP %d | %s\n", selected.ID, selected.Health, statusLabel(selected.Status))
		fmt.Fprintf(&b, "Run %s | Avg %s | Boss %s\n", format.Duration(selected.TotalTime()), format.Duration(selected.AverageFloorClearTime()), format.Duration(selected.Dungeon.BossKillTime))
	} else if m.selectedID != 0 {
		fmt.Fprintf(&b, "Inspect P%d | not registered here\n", m.selectedID)
	} else {
		b.WriteString("Inspect: click a player row\n")
	}
	b.WriteString("Map: START > F... > BOSS > EXIT\n")
	b.WriteByte('\n')

	rooms := make([][]string, 0, int(game.Cfg.Floors)+2)
	startPlayers := m.playersAtStart()
	rooms = append(rooms, dungeonSpecialRoom("START", startPlayers, contentWidth, strings.Contains(startPlayers, "[@")))
	for i, floor := range m.mapFloors(selected, hasSelected) {
		rooms = append(rooms, dungeonRoom(m, floor, uint8(i), contentWidth, hasSelected))
	}
	exitPlayers := m.playersAtExit()
	rooms = append(rooms, dungeonSpecialRoom("EXIT", exitPlayers, contentWidth, strings.Contains(exitPlayers, "[@")))

	for i, lines := range rooms {
		for _, line := range lines {
			b.WriteString(line)
			b.WriteByte('\n')
		}
		if i < len(rooms)-1 {
			connector := centerASCII("||", contentWidth)
			b.WriteString(connector)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) mapFloors(selected player.Player, hasSelected bool) []player.Floor {
	if hasSelected {
		return selected.Dungeon.Floors
	}
	return player.New(0, game.Cfg.Floors, game.Cfg.Monsters).Dungeon.Floors
}

func dungeonRoom(m Model, floor player.Floor, floorID uint8, width int, hasSelected bool) []string {
	innerWidth := width - 2
	if innerWidth < 18 {
		innerWidth = 18
	}
	border := "+" + strings.Repeat("-", innerWidth) + "+"
	roomType := "FLOOR"
	threat := monstersASCII(floor.MonstersLeft)
	if floor.IsBoss {
		roomType = "BOSS"
		threat = "BOSS"
	}
	cleared := "open"
	if floor.Cleared {
		cleared = "cleared"
	}

	prefix := " "
	if hasSelected && m.selectedID != 0 {
		if selected, ok := player.Players[m.selectedID]; ok && selected.EnteredDungeon && !selected.Finished && selected.Floor == floorID {
			prefix = fmt.Sprintf("@P%d", selected.ID)
		}
	}
	playersHere := m.playersOnFloor(floorID)

	lines := []string{
		border,
		asciiLine(fmt.Sprintf("%s %02d %-5s %s", prefix, floorID, roomType, cleared), innerWidth),
		asciiLine(fmt.Sprintf("%s T:%s P:%s", threat, format.Duration(floor.TimeSpent), playersHere), innerWidth),
		border,
	}
	if floor.IsBoss {
		lines = styleRoom(lines, bossStyle)
	} else if floor.Cleared {
		lines = styleRoom(lines, clearedStyle)
	}
	if hasSelected && prefix != " " {
		lines = styleRoom(lines, selectedRoomStyle)
	}
	return lines
}

func dungeonSpecialRoom(label, players string, width int, highlighted bool) []string {
	innerWidth := width - 2
	if innerWidth < 18 {
		innerWidth = 18
	}
	border := "+" + strings.Repeat("=", innerWidth) + "+"
	lines := []string{
		border,
		asciiLine(fmt.Sprintf("%s | players:%s", label, players), innerWidth),
		border,
	}
	if highlighted {
		return styleRoom(lines, selectedRoomStyle)
	}
	return lines
}

func styleRoom(lines []string, style lipgloss.Style) []string {
	styled := make([]string, len(lines))
	for i, line := range lines {
		styled[i] = style.Render(line)
	}
	return styled
}

func monstersASCII(monstersLeft uint8) string {
	if monstersLeft == 0 {
		return "M:clear"
	}
	visible := min(int(monstersLeft), 8)
	monsters := strings.Repeat("M", visible)
	if int(monstersLeft) > visible {
		monsters += "+"
	}
	return "M:" + monsters
}

func asciiLine(text string, width int) string {
	return "|" + padASCII(text, width) + "|"
}

func padASCII(text string, width int) string {
	if len(text) > width {
		if width <= 1 {
			return text[:width]
		}
		return text[:width-1] + "~"
	}
	return text + strings.Repeat(" ", width-len(text))
}

func centerASCII(text string, width int) string {
	if len(text) >= width {
		return text
	}
	left := (width - len(text)) / 2
	return strings.Repeat(" ", left) + text
}

func (m Model) playersView(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Players"))
	b.WriteByte('\n')
	ids := playerIDs()
	if len(ids) == 0 {
		b.WriteString(mutedStyle.Render("No players yet"))
		return b.String()
	}
	b.WriteString(mutedStyle.Render("click a row to inspect"))
	b.WriteByte('\n')
	b.WriteString("ID HP  St  Fl In Dn\n")
	limit := max(height-4, 1)
	for i, id := range ids {
		if i >= limit {
			b.WriteString("...")
			break
		}
		p := player.Players[id]
		line := fmt.Sprintf("%-2d %-3d %-3s %-2d %-2s %-2s", p.ID, p.Health, shortStatus(p.Status), p.Floor, boolLabel(p.EnteredDungeon), boolLabel(p.Finished))
		switch p.ID {
		case m.selectedID:
			line = selectedPlayerStyle.Render(line)
		case m.activeID:
			line = currentStyle.Render(line)
		}
		b.WriteString(zone.Mark(m.playerZone(id), line))
		if i < len(ids)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) playerZone(id uint8) string {
	return fmt.Sprintf("%splayer-%d", m.zoneID, id)
}

func (m Model) logView(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Output"))
	b.WriteByte('\n')
	visible := max(height-2, 1)
	start := max(len(m.Log)-visible, 0)
	for i := start; i < len(m.Log); i++ {
		b.WriteString(m.Log[i])
		if i < len(m.Log)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) activePlayer() (player.Player, bool) {
	if m.selectedID != 0 {
		if p, ok := player.Players[m.selectedID]; ok {
			return p, true
		}
	}
	if m.activeID != 0 {
		if p, ok := player.Players[m.activeID]; ok {
			return p, true
		}
	}
	ids := playerIDs()
	if len(ids) == 0 {
		return player.Player{}, false
	}
	return player.Players[ids[0]], true
}

func (m Model) selectedPlayer() (player.Player, bool) {
	if m.selectedID != 0 {
		p, ok := player.Players[m.selectedID]
		return p, ok
	}
	return m.activePlayer()
}

func (m Model) playersAtStart() string {
	var ids []uint8
	for _, id := range playerIDs() {
		p := player.Players[id]
		if !p.EnteredDungeon && !p.Finished {
			ids = append(ids, id)
		}
	}
	return playerMarkers(ids, m.selectedID)
}

func (m Model) playersAtExit() string {
	var ids []uint8
	for _, id := range playerIDs() {
		p := player.Players[id]
		if p.Finished {
			ids = append(ids, id)
		}
	}
	return playerMarkers(ids, m.selectedID)
}

func (m Model) playersOnFloor(floorID uint8) string {
	var ids []uint8
	for _, id := range playerIDs() {
		p := player.Players[id]
		if p.EnteredDungeon && !p.Finished && p.Floor == floorID {
			ids = append(ids, id)
		}
	}
	return playerMarkers(ids, m.selectedID)
}

func playerMarkers(ids []uint8, selectedID uint8) string {
	if len(ids) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == selectedID {
			parts = append(parts, fmt.Sprintf("[@%d]", id))
			continue
		}
		parts = append(parts, fmt.Sprintf("@%d", id))
	}
	return strings.Join(parts, " ")
}

func playerIDs() []uint8 {
	ids := make([]uint8, 0, len(player.Players))
	for id := range player.Players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func statusLabel(status player.Status) string {
	switch status {
	case player.StatusSuccess:
		return "success"
	case player.StatusFail:
		return "fail"
	case player.StatusDisqual:
		return "disqual"
	default:
		return "unknown"
	}
}

func shortStatus(status player.Status) string {
	switch status {
	case player.StatusSuccess:
		return "ok"
	case player.StatusFail:
		return "bad"
	case player.StatusDisqual:
		return "dq"
	default:
		return "?"
	}
}

func boolLabel(v bool) string {
	if v {
		return "y"
	}
	return "-"
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("238")).
			Padding(0, 1)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
	playersPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 2, 0, 1)
	logStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("244")).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("25"))
	selectedPlayerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("63")).
				Bold(true)
	selectedRoomStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("24")).
				Bold(true)
	bossStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")).
			Bold(true)
	clearedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78"))
	doneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236")).
			Padding(1, 2)
)

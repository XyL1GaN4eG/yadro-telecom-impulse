package tui

import (
	"bufio"
	"fmt"
	"impulse/internal/event"
	"impulse/internal/format"
	"impulse/internal/game"
	"impulse/internal/parser"
	"impulse/internal/player"
	"impulse/internal/replay"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const tickEvery = 500 * time.Millisecond

type Entry struct {
	Line string
	Cmd  event.Command
	Err  error
}

type Model struct {
	Entries  []Entry
	Index    int
	Log      []string
	Playing  bool
	LastErr  string
	Width    int
	Height   int
	activeID uint8
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

	_, err = tea.NewProgram(
		NewModel(entries),
		tea.WithAltScreen(),
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
	defer file.Close()

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
	return Model{
		Entries: entries,
		Width:   100,
		Height:  32,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
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
		case " ":
			m.Playing = !m.Playing
			if m.Playing {
				return m, tick()
			}
		case "right", "n":
			m.Step()
		case "left", "p":
			m.StepBack()
		case "r":
			m.Reset()
		case "g":
			m.PlayToEnd()
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

	m.activeID = entry.Cmd.PlayerID
	lines := replay.ProcessCommand(entry.Cmd)
	if len(lines) == 0 {
		m.LastErr = "event ignored"
	}
	m.Log = append(m.Log, lines...)
	m.Index++
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
	m.Playing = false
	for m.Index < target {
		m.Step()
	}
}

func (m Model) View() string {
	width := max(m.Width, 80)
	height := max(m.Height, 24)
	panelHeight := max(height-9, 10)
	timelineWidth := clamp(width/4, 24, 36)
	playersWidth := clamp(width/4, 26, 38)
	centerWidth := max(width-timelineWidth-playersWidth-4, 24)
	logHeight := max(height-panelHeight-4, 5)

	top := headerStyle.Width(width).Render(m.header())
	timeline := panelStyle.Width(timelineWidth).Height(panelHeight).Render(m.timelineView(panelHeight))
	dungeon := panelStyle.Width(centerWidth).Height(panelHeight).Render(m.dungeonView(panelHeight))
	players := panelStyle.Width(playersWidth).Height(panelHeight).Render(m.playersView(panelHeight))
	log := logStyle.Width(width).Height(logHeight).Render(m.logView(logHeight))
	return lipgloss.JoinVertical(lipgloss.Left, top, lipgloss.JoinHorizontal(lipgloss.Top, timeline, dungeon, players), log)
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
		"Floors %d | Monsters %d | Open %s | Duration %dh | Step %d/%d | %s%s",
		game.Cfg.Floors,
		game.Cfg.Monsters,
		game.Cfg.OpenAt,
		game.Cfg.Duration,
		m.Index,
		len(m.Entries),
		mode,
		errText,
	)
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

func (m Model) dungeonView(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Dungeon"))
	b.WriteByte('\n')
	p, ok := m.activePlayer()
	if !ok {
		b.WriteString("No active player yet")
		return b.String()
	}
	b.WriteString(fmt.Sprintf("Player %d | HP %d | %s\n\n", p.ID, p.Health, statusLabel(p.Status)))
	limit := max(height-5, 1)
	for i, floor := range p.Dungeon.Floors {
		if i >= limit {
			b.WriteString("...")
			break
		}
		marker := " "
		if uint8(i) == p.Floor {
			marker = ">"
		}
		kind := "floor"
		if floor.IsBoss {
			kind = "boss"
		}
		cleared := "open"
		if floor.Cleared {
			cleared = "cleared"
		}
		line := fmt.Sprintf("%s %02d %-5s monsters:%d %s", marker, i, kind, floor.MonstersLeft, cleared)
		if uint8(i) == p.Floor {
			line = currentStyle.Render(line)
		}
		b.WriteString(line)
		if i < len(p.Dungeon.Floors)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) playersView(height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Players"))
	b.WriteByte('\n')
	ids := playerIDs()
	if len(ids) == 0 {
		b.WriteString("No players")
		return b.String()
	}
	b.WriteString("ID  HP  Status    Fl  In  Done\n")
	limit := max(height-3, 1)
	for i, id := range ids {
		if i >= limit {
			b.WriteString("...")
			break
		}
		p := player.Players[id]
		line := fmt.Sprintf("%-3d %-3d %-9s %-3d %-3t %-4t", p.ID, p.Health, statusLabel(p.Status), p.Floor, p.EnteredDungeon, p.Finished)
		if p.ID == m.activeID {
			line = currentStyle.Render(line)
		}
		b.WriteString(line)
		if i < len(ids)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
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
	logStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("244")).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("25"))
	doneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

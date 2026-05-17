Output docs/assets/impulse-tui.gif

Require go
Require starship

Set Shell "fish"
Set FontFamily "Hack Nerd Font Mono"
Set FontSize 18
Set Width 1320
Set Height 820
Set Padding 22
Set Theme "Dracula"
Set TypingSpeed 34ms
Set PlaybackSpeed 0.85
Set Framerate 20
Env TERM xterm-256color
Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"

Hide
Type "source docs/vhs/prompt.fish"
Enter
Sleep 500ms
Ctrl+L
Sleep 300ms
Show

Type "go run . tui docs/config.json docs/events"
Sleep 500ms
Enter
Sleep 1s
Space
Sleep 15s
Type "q"
Sleep 300ms

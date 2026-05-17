Output docs/assets/impulse-replay.gif

Require go
Require starship

Set Shell "fish"
Set FontFamily "Hack Nerd Font Mono"
Set FontSize 20
Set Width 1180
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 28ms
Env TERM xterm-256color

Hide
Type "source docs/vhs/prompt.fish"
Enter
Sleep 500ms
Ctrl+L
Sleep 300ms
Show

Type "cat docs/events | go run . docs/config.json"
Sleep 300ms
Enter
Sleep 5s

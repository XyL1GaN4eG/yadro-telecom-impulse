Output docs/assets/impulse-tests.gif

Require go
Require starship

Set Shell "fish"
Set FontFamily "Hack Nerd Font Mono"
Set FontSize 23
Set Width 1180
Set Height 640
Set Padding 24
Set Theme "Dracula"
Set TypingSpeed 34ms
Set PlaybackSpeed 0.9
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

Type "go test ./..."
Sleep 300ms
Enter
Sleep 4s

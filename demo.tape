Output demo.gif
Set Shell bash
Set Width 1400
Set Height 700
Set FontSize 14
Set Padding 20
Set Theme "Catppuccin Mocha"

Type "./bin/cctv --claude-dir /tmp/cctv-demo"
Enter
Sleep 2s

# Navigate down through sessions
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# Show stats popup
Type "s"
Sleep 3s
Escape
Sleep 1s

# Filter by project
Type "/"
Sleep 500ms
Type "project:backend$"
Sleep 2s
Escape
Sleep 1s

# Clear filter and show all
Type "/"
Sleep 300ms
Ctrl+U
Sleep 300ms
Enter
Sleep 1s

# Show help
Type "?"
Sleep 2s

# Quit
Type "q"
Sleep 500ms

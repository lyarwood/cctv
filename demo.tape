Output demo.gif
Require ./bin/cctv
Set Shell bash
Set Width 1400
Set Height 700
Set FontSize 14
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# demo/ contains the fake session data; the Makefile stages it into
# the directory layout cctv expects before running this tape.
# A fake claude script is placed in /tmp/cctv-demo/bin/ so resume works.

# --- TUI ---
# Use the fake claude from the demo staging area
Type "PATH=/tmp/cctv-demo/bin:$PATH ./bin/cctv --claude-dir /tmp/cctv-demo"
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

# Filter by project (regex) - Enter to apply
Type "/"
Sleep 500ms
Type "project:backend$"
Enter
Sleep 2s

# Filter by PR - clear previous, type new, Enter to apply
Type "/"
Sleep 300ms
Ctrl+U
Sleep 300ms
Type "pr:webapp"
Enter
Sleep 2s

# Clear filter and show all
Type "/"
Sleep 300ms
Ctrl+U
Sleep 300ms
Enter
Sleep 1s

# Resume a session - navigate to the API review session
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 3s
# Press a key to dismiss the fake claude session and return to cctv
Type " "
Sleep 1s

# Show help after returning from resume
Type "?"
Sleep 2s

# Quit
Type "q"
Sleep 1s

# --- CLI ---
Type "./bin/cctv list --claude-dir /tmp/cctv-demo"
Enter
Sleep 2s

Type "./bin/cctv list --claude-dir /tmp/cctv-demo --pr webapp"
Enter
Sleep 2s

Type "./bin/cctv list --claude-dir /tmp/cctv-demo --project backend$ --json | head -20"
Enter
Sleep 3s

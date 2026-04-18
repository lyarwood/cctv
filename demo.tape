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

# Put fake claude on PATH
Hide
Type "export PATH=/tmp/cctv-demo/bin:$PATH"
Enter
Sleep 500ms
Show

# --- TUI ---
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

# Filter by project (regex)
Type "/"
Sleep 500ms
Type "project:backend$"
Sleep 2s
Escape
Sleep 1s

# Filter by PR
Type "/"
Sleep 300ms
Ctrl+U
Sleep 300ms
Type "pr:webapp"
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

# Resume a session - navigate to the API review session
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 4s

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

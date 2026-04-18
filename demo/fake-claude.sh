#!/usr/bin/env bash
# Fake claude binary for demo recordings.
# Prints a mock session header and waits briefly to simulate resuming.
SESSION_ID="${3:-unknown}"
clear
echo ""
echo "  ╭─────────────────────────────────────────────────────────╮"
echo "  │  Claude Code v2.1.112                                   │"
echo "  │                                                         │"
echo "  │  Resuming session ${SESSION_ID:0:8}...                  │"
echo "  │                                                         │"
echo "  │  Project: /home/dev/webapp                              │"
echo "  │  Branch:  feature/api-v2                                │"
echo "  │  Messages: 18                                           │"
echo "  │                                                         │"
echo "  │  > Review and address PR feedback on the API endpoints  │"
echo "  │                                                         │"
echo "  ╰─────────────────────────────────────────────────────────╯"
echo ""
echo "  Session restored. Type your prompt or /help for commands."
echo ""
sleep 3

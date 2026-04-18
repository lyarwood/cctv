#!/usr/bin/env bash
# Fake claude binary for demo recordings.
# Prints a mock session header and waits briefly to simulate resuming.
# Avoids `clear` which interferes with VHS terminal recording.
SESSION_ID="${3:-unknown}"

printf '\n\n'
printf '  \033[1;35m╭─────────────────────────────────────────────────────────╮\033[0m\n'
printf '  \033[1;35m│\033[0m  \033[1mClaude Code\033[0m v2.1.112                                   \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m                                                         \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m  Resuming session \033[33m%.8s\033[0m...                          \033[1;35m│\033[0m\n' "$SESSION_ID"
printf '  \033[1;35m│\033[0m                                                         \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m  Project: \033[36m/tmp/cctv-demo/webapp\033[0m                              \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m  Branch:  \033[32mfeature/api-v2\033[0m                                \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m  Messages: 18                                           \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m                                                         \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m  > Review and address PR feedback on the API endpoints  \033[1;35m│\033[0m\n'
printf '  \033[1;35m│\033[0m                                                         \033[1;35m│\033[0m\n'
printf '  \033[1;35m╰─────────────────────────────────────────────────────────╯\033[0m\n'
printf '\n'
printf '  \033[2mSession restored. Press any key to return to cctv.\033[0m\n\n'

read -rsn1

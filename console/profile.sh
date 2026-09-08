# Sourced by the interactive shell opened from Unraid's "Console" button (busybox ash reads $ENV).
export PS1='unraid-gateway:\w$ '
alias ll='ls -la'
if [ -t 1 ]; then gw help; echo; gw status; echo; echo "  Type 'gw' for the list of commands."; echo; fi

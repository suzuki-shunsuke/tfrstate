#!/usr/bin/env bash

set -eu
set -o pipefail

cd "$(dirname "$0")/.."

command_console() {
  echo '```console'
  echo "$ $*"
  "$@"
  echo '```'
}

commands() {
  for cmd in add completion; do
    echo "
## tf-remote-state-find $cmd

$(command_console tf-remote-state-find help $cmd)"
  done
}

echo "# Usage

<!-- This is generated by scripts/generate-usage.sh. Don't edit this file directly. -->

$(command_console tf-remote-state-find help)
$(commands)
" > USAGE.md

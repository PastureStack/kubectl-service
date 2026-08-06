#!/bin/bash
set -euo pipefail

token=${KUBECTL_SHELL_TOKEN:-${1:-}}

if [ -z "${token}" ]; then
    echo 'No token provided' >&2
    exit 1
fi

echo "# Run kubectl commands inside here"
echo "# e.g. kubectl get all"
export TERM=screen-256color

set --
exec env KUBECTL_SHELL_TOKEN="${token}" /usr/bin/shell-setup.sh

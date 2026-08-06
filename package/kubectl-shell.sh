#!/bin/bash
set -euo pipefail

token=${1:-}

if [ -z "${token}" ]; then
    echo No token provided
    exit 1
fi

echo "# Run kubectl commands inside here"
echo "# e.g. kubectl get all"
export TERM=screen-256color

KUBECTL_SHELL_TOKEN="${token}" unshare --fork --pid --mount-proc --mount shell-setup.sh

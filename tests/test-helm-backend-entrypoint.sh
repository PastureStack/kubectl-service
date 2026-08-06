#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() {
    rm -rf -- "$temporary_directory"
}
trap cleanup EXIT

# shellcheck disable=SC1090
source "$repo_root/package/kubectl-service.sh"

call_log="$temporary_directory/calls"
HOME="$temporary_directory/home"
mkdir -p "$HOME"

chown() {
    printf 'chown %s\n' "$*" >> "$call_log"
}
run_helm2() {
    printf 'helm2 %s\n' "$*" >> "$call_log"
}
run_helm4() {
    printf 'helm4 %s\n' "$*" >> "$call_log"
}

PASTURESTACK_HELM_BACKEND=legacy-helm2
validate_helm_backend
initialize_helm_backend
grep -Fqx 'helm2 version --client' "$call_log"
grep -Fqx 'helm2 init -c' "$call_log"
if grep -Fq 'helm4 ' "$call_log"; then
    echo 'legacy initialization unexpectedly called Helm 4' >&2
    exit 1
fi

: > "$call_log"
PASTURESTACK_HELM_BACKEND=helm4
validate_helm_backend
initialize_helm_backend
grep -Fqx 'helm4 version --short' "$call_log"
test -d "$HELM_CACHE_HOME"
test -d "$HELM_CONFIG_HOME"
test -d "$HELM_DATA_HOME"
if grep -Fq 'helm2 ' "$call_log"; then
    echo 'Helm 4 initialization unexpectedly called Helm 2' >&2
    exit 1
fi

PASTURESTACK_HELM_BACKEND=auto
if validate_helm_backend >/dev/null 2>&1; then
    echo 'unsupported Helm backend unexpectedly passed validation' >&2
    exit 1
fi

PASTURESTACK_HELM_BACKEND=''
if validate_helm_backend >/dev/null 2>&1; then
    echo 'empty Helm backend unexpectedly passed explicit validation' >&2
    exit 1
fi

echo 'KUBECTL_SERVICE_HELM_BACKEND_ENTRYPOINT_TEST_OK'

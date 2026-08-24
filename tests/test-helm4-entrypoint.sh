#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() { rm -rf -- "$temporary_directory"; }
trap cleanup EXIT

# shellcheck disable=SC1090
source "$repo_root/package/kubectl-service.sh"

HOME="$temporary_directory/home"
mkdir -p "$HOME"
SERVER='https://kubernetes.example.invalid:6443/path?value="quoted"'
unset KUBECONFIG
write_kubeconfig
test "$KUBECONFIG" = "$HOME/.kube/config"
test "$(stat -c '%a' "$HOME/.kube")" = 700
test "$(stat -c '%a' "$HOME/.kube/config")" = 600
grep -Fqx '    server: "https://kubernetes.example.invalid:6443/path?value=\"quoted\""' "$HOME/.kube/config"

SERVER=$'https://kubernetes.example.invalid:6443/\ninvalid'
if write_kubeconfig >/dev/null 2>&1; then
    echo 'SERVER containing a newline unexpectedly produced a kubeconfig' >&2
    exit 1
fi

SERVER='https://user:password@kubernetes.example.invalid:6443/'
if write_kubeconfig >/dev/null 2>&1; then
    echo 'SERVER containing embedded credentials unexpectedly produced a kubeconfig' >&2
    exit 1
fi

grep -Fq '/usr/bin/helm version --short' "$repo_root/package/kubectl-service.sh"
if grep -R -i -E 'helm[ -]?2|tiller|legacy-helm' "$repo_root/main.go" "$repo_root/helm" "$repo_root/package"; then
    echo 'Removed Helm 2/Tiller runtime path remains' >&2
    exit 1
fi

echo 'KUBECTL_SERVICE_HELM4_ENTRYPOINT_TEST_OK'

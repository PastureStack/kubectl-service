#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() { rm -rf -- "$temporary_directory"; }
trap cleanup EXIT

system_bundle="$temporary_directory/system-ca.pem"
platform_ca="$temporary_directory/platform-ca.pem"
home_directory="$temporary_directory/home"
output_bundle="$home_directory/.local/share/ca-certificates/ca-bundle.crt"
mkdir -p "$home_directory"
printf '%s\n' '-----BEGIN CERTIFICATE-----' 'SYSTEM-CA-TEST-FIXTURE' '-----END CERTIFICATE-----' > "$system_bundle"
printf '%s\n' '-----BEGIN CERTIFICATE-----' 'PLATFORM-CA-TEST-FIXTURE' '-----END CERTIFICATE-----' > "$platform_ca"

HOME="$home_directory" SYSTEM_CA_BUNDLE="$system_bundle" PLATFORM_CA_ROOT="$platform_ca" \
SSL_CERT_FILE="$output_bundle" bash "$repo_root/package/update-platform-ca" "$output_bundle"
test "$(stat -c '%a' "$output_bundle")" = 600
grep -Fqx 'SYSTEM-CA-TEST-FIXTURE' "$output_bundle"
grep -Fqx 'PLATFORM-CA-TEST-FIXTURE' "$output_bundle"

test "$(awk '$1 == "USER" { user=$2 } END { print user }' "$repo_root/package/Dockerfile")" = '65534:65534'
grep -Fq 'io.pasturestack.runtime.user="65534:65534"' "$repo_root/package/Dockerfile"
grep -Fq 'io.pasturestack.shell.requires-privileged="false"' "$repo_root/package/Dockerfile"
grep -Fq 'ARG KUBERNETES_VERSION=v1.36.4' "$repo_root/package/Dockerfile"
grep -Fq 'ARG HELM_VERSION=v4.2.4' "$repo_root/package/Dockerfile"
grep -Fq 'KUBERNETES_SOURCE_SHA256=3c28f11492472df48e658551bf268fd92938b127b0f9dcef7090ac800318c821' "$repo_root/package/Dockerfile"
grep -Fq 'HELM_SOURCE_SHA256=f1a1aa56fff071cfa1eecc18c6a7212b2e5c7a0a123fd8e4dcae620bd51b4286' "$repo_root/package/Dockerfile"
grep -Fq 'HELM_ORAS_GO_VERSION=2.6.2' "$repo_root/package/Dockerfile"
test "$(grep -c 'golang:1.27.0-bookworm@sha256:484ef6066fa69acb059fdfeda7ba2b8f7391f2ef6abc6f9b8411e669ebd56466' "$repo_root/package/Dockerfile")" -eq 2
grep -Fq -- '--cap-drop ALL' "$repo_root/scripts/package"
grep -Fq -- '--security-opt no-new-privileges:true' "$repo_root/scripts/package"

if grep -R -i -E 'helm[ -]?2|tiller|legacy-helm|v1\.12\.' "$repo_root/main.go" "$repo_root/helm" "$repo_root/package" "$repo_root/scripts/package"; then
    echo 'Removed EOL runtime marker remains in an active path' >&2
    exit 1
fi

echo 'KUBECTL_SERVICE_NON_ROOT_RUNTIME_TEST_OK user=65534:65534'

#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() {
    rm -rf -- "$temporary_directory"
}
trap cleanup EXIT

mkdir -p "$temporary_directory/scripts" "$temporary_directory/migration"
cp "$repo_root/scripts/check-migration-targets" "$temporary_directory/scripts/"
cp "$repo_root/migration/targets.lock.env" "$temporary_directory/migration/"

bash "$temporary_directory/scripts/check-migration-targets" --offline >/dev/null

expect_failure() {
    if "$@" >/dev/null 2>&1; then
        echo "command unexpectedly succeeded: $*" >&2
        exit 1
    fi
}

sed -i 's/^HELM3_BRIDGE_GIT_COMMIT=.*/HELM3_BRIDGE_GIT_COMMIT=short/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/migration/targets.lock.env" "$temporary_directory/migration/targets.lock.env"

sed -i 's/^HELM4_TARGET_LINUX_AMD64_SHA256=.*/HELM4_TARGET_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/migration/targets.lock.env" "$temporary_directory/migration/targets.lock.env"

sed -i 's/^HELM4_TARGET_VERSION=.*/HELM4_TARGET_VERSION=v3.21.3/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --unknown

echo "KUBECTL_SERVICE_MIGRATION_TARGETS_NEGATIVE_TEST_OK"

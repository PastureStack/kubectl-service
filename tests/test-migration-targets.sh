#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() {
    rm -rf -- "$temporary_directory"
}
trap cleanup EXIT

mkdir -p "$temporary_directory/scripts" "$temporary_directory/migration" "$temporary_directory/package"
cp "$repo_root/scripts/check-migration-targets" "$temporary_directory/scripts/"
cp "$repo_root/migration/targets.lock.env" "$temporary_directory/migration/"
cp "$repo_root/migration/kubernetes-upgrade-path.tsv" "$temporary_directory/migration/"
cp "$repo_root/package/helm-v4.2.3-oras-go-v2.6.2.patch" "$temporary_directory/package/"
cp "$repo_root/package/helm-v3.21.3-security.patch" "$temporary_directory/package/"
cp "$repo_root/package/helm-2to3-v0.11.0-security.patch" "$temporary_directory/package/"

reset_locks() {
    cp "$repo_root/migration/targets.lock.env" "$temporary_directory/migration/targets.lock.env"
    cp "$repo_root/migration/kubernetes-upgrade-path.tsv" "$temporary_directory/migration/kubernetes-upgrade-path.tsv"
}

bash -n "$temporary_directory/scripts/check-migration-targets"
bash "$temporary_directory/scripts/check-migration-targets" --offline >/dev/null

expect_failure() {
    if "$@" >/dev/null 2>&1; then
        echo "command unexpectedly succeeded: $*" >&2
        exit 1
    fi
}

sed -i 's/^HELM3_BRIDGE_GIT_COMMIT=.*/HELM3_BRIDGE_GIT_COMMIT=short/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM_BUILD_VERSION=.*/HELM_BUILD_VERSION=v2.17.0-1/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256=.*/HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^TILLER_REBUILT_LINUX_AMD64_SHA256=.*/TILLER_REBUILT_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM3_BRIDGE_SOURCE_SHA256=.*/HELM3_BRIDGE_SOURCE_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=.*/HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM3_BRIDGE_BUILD_VERSION=.*/HELM3_BRIDGE_BUILD_VERSION=v3.21.3-1/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

printf '\n# corrupted test fixture\n' >> "$temporary_directory/package/helm-v3.21.3-security.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-v3.21.3-security.patch" "$temporary_directory/package/"

rm "$temporary_directory/package/helm-v3.21.3-security.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-v3.21.3-security.patch" "$temporary_directory/package/"

sed -i 's/^HELM2TO3_APPLIED_PATCH_SHA256=.*/HELM2TO3_APPLIED_PATCH_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM2TO3_REBUILT_LINUX_AMD64_SHA256=.*/HELM2TO3_REBUILT_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM2TO3_BUILD_VERSION=.*/HELM2TO3_BUILD_VERSION=0.11.1/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM2TO3_HELM3_VERSION=.*/HELM2TO3_HELM3_VERSION=v3.21.2/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

printf '\n# corrupted test fixture\n' >> "$temporary_directory/package/helm-2to3-v0.11.0-security.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-2to3-v0.11.0-security.patch" "$temporary_directory/package/"

rm "$temporary_directory/package/helm-2to3-v0.11.0-security.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-2to3-v0.11.0-security.patch" "$temporary_directory/package/"

sed -i 's/^HELM4_TARGET_LINUX_AMD64_SHA256=.*/HELM4_TARGET_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_TARGET_LICENSE_SHA256=.*/HELM4_TARGET_LICENSE_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_TARGET_SOURCE_SHA256=.*/HELM4_TARGET_SOURCE_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_SECURITY_PATCH_COMMIT=.*/HELM4_SECURITY_PATCH_COMMIT=short/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_SECURITY_PATCH_SHA256=.*/HELM4_SECURITY_PATCH_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_APPLIED_PATCH_SHA256=.*/HELM4_APPLIED_PATCH_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_ORAS_GO_VERSION=.*/HELM4_ORAS_GO_VERSION=2.6.2/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_BUILD_VERSION=.*/HELM4_BUILD_VERSION=v4.2.4+1/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^HELM4_REBUILT_LINUX_AMD64_SHA256=.*/HELM4_REBUILT_LINUX_AMD64_SHA256=invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

printf '\n# corrupted test fixture\n' >> "$temporary_directory/package/helm-v4.2.3-oras-go-v2.6.2.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-v4.2.3-oras-go-v2.6.2.patch" "$temporary_directory/package/"

rm "$temporary_directory/package/helm-v4.2.3-oras-go-v2.6.2.patch"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
cp "$repo_root/package/helm-v4.2.3-oras-go-v2.6.2.patch" "$temporary_directory/package/"

sed -i 's/^HELM4_TARGET_VERSION=.*/HELM4_TARGET_VERSION=v3.21.3/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i '/^20\t/d' "$temporary_directory/migration/kubernetes-upgrade-path.tsv"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

awk 'BEGIN {FS=OFS="\t"} $1 == 20 {$4="invalid"} {print}' \
    "$temporary_directory/migration/kubernetes-upgrade-path.tsv" > "$temporary_directory/migration/path.tmp"
mv "$temporary_directory/migration/path.tmp" "$temporary_directory/migration/kubernetes-upgrade-path.tsv"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

awk 'BEGIN {FS=OFS="\t"} $1 == 33 {$5="SUPPORTED"} {print}' \
    "$temporary_directory/migration/kubernetes-upgrade-path.tsv" > "$temporary_directory/migration/path.tmp"
mv "$temporary_directory/migration/path.tmp" "$temporary_directory/migration/kubernetes-upgrade-path.tsv"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^KUBERNETES_UPGRADE_STEP_COUNT=.*/KUBERNETES_UPGRADE_STEP_COUNT=23/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^KUBERNETES_TARGET_VERSION=.*/KUBERNETES_TARGET_VERSION=v1.36.2/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

sed -i 's/^KUBERNETES_SOURCE_VERSION=.*/KUBERNETES_SOURCE_VERSION=v1.12.10-invalid/' "$temporary_directory/migration/targets.lock.env"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

rm "$temporary_directory/migration/kubernetes-upgrade-path.tsv"
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --offline
reset_locks

expect_failure bash "$temporary_directory/scripts/check-migration-targets" --unknown
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --download-helm extra
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --download-kubernetes
expect_failure bash "$temporary_directory/scripts/check-migration-targets" --download-kubernetes invalid

echo "KUBECTL_SERVICE_MIGRATION_TARGETS_NEGATIVE_TEST_OK"

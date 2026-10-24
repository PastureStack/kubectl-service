#!/usr/bin/env bash
# Host-side gate: no additional interpreter is installed in either product image.
set -euo pipefail
test -n "${DAPPER_IMAGE:?}"
test -d "${DAPPER_TRIVY_CACHE:?}"
test -s dist/images
test "$(wc -l < dist/images)" -eq 1
image=$(cat dist/images)
[[ "$image" =~ ^pasturestack/kubectl-service:v[0-9]+\.[0-9]+\.[0-9]+$ ]]
socket_group=$(stat -c '%g' /var/run/docker.sock)
run_scan() {
    docker run --rm --user "$(id -u):$(id -g)" --group-add "$socket_group" \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$DAPPER_TRIVY_CACHE:/tmp/trivy-cache" \
        -v "$PWD:/workspace:ro" -v "$PWD/dist:/evidence" \
        --entrypoint /usr/local/bin/trivy "$DAPPER_IMAGE" "$@"
}

for tool in HELM KUBECTL; do
    docker run --rm --entrypoint /usr/bin/cat "$image" "/licenses/$tool-PACKAGES.txt" > "dist/$tool-PACKAGES.txt"
    test -s "dist/$tool-PACKAGES.txt"
    if grep -Eq '^golang.org/x/crypto/openpgp(/|$)' "dist/$tool-PACKAGES.txt"; then echo 'Forbidden OpenPGP dependency' >&2; exit 1; fi
done
for tool in DOCKER BUILDX TRIVY; do
    docker run --rm --entrypoint /usr/bin/cat "$DAPPER_IMAGE" "/licenses/$tool-PACKAGES.txt" > "dist/$tool-PACKAGES.txt"
    test -s "dist/$tool-PACKAGES.txt"
    if grep -Eq '^golang.org/x/crypto/openpgp(/|$)' "dist/$tool-PACKAGES.txt"; then echo 'Forbidden OpenPGP dependency' >&2; exit 1; fi
done

for binary in helm kubectl kubectl-service; do
    jq -e --arg binary "$binary" 'any(.Results[]?;
      .Type == "gobinary" and (.Target | split("/") | last) == $binary
      and (.Packages | length > 0))' dist/kubectl-service.security-raw.json >/dev/null
done
bundle_image='ghcr.io/pasturestack/ubuntu-security-backports:v0.1.0@sha256:2742a3d429e4e14bc587f6dedc51ded4246cf90bbaedcd86c67d0885e32a6f2a'
mkdir -p dist/bundle-tools
bundle_container=$(docker create "$bundle_image" unused)
if docker cp "$bundle_container:/out/verify-consumer-evidence.py" dist/bundle-tools/verify-consumer-evidence.py && \
    docker cp "$bundle_container:/out/SHA256SUMS" dist/bundle-tools/SHA256SUMS; then
    docker rm "$bundle_container" >/dev/null
else
    docker rm "$bundle_container" >/dev/null
    exit 1
fi
(cd dist/bundle-tools && \
    test "$(grep -Ec '^[0-9a-f]{64}  verify-consumer-evidence\.py$' SHA256SUMS)" -eq 1 && \
    grep -E '^[0-9a-f]{64}  verify-consumer-evidence\.py$' SHA256SUMS | sha256sum --check -)
python3 dist/bundle-tools/verify-consumer-evidence.py \
    --evidence-dir dist/ubuntu-backports --sbom dist/kubectl-service.cdx.json \
    --bundle-checksums dist/bundle-tools/SHA256SUMS \
    --expected-bundle-digest 'sha256:2742a3d429e4e14bc587f6dedc51ded4246cf90bbaedcd86c67d0885e32a6f2a' > dist/ubuntu-backports-validation.json
tar -cf dist/ubuntu-backports-evidence.tar -C dist ubuntu-backports ubuntu-backports-validation.json
python3 .github/scripts/enforce-runtime-vulnerabilities.py --phase runtime --repository "$GITHUB_REPOSITORY"
run_scan fs --cache-dir /tmp/trivy-cache --skip-db-update --scanners secret --exit-code 1 /workspace
run_scan image --cache-dir /tmp/trivy-cache --skip-db-update --offline-scan \
    --scanners vuln,secret --list-all-pkgs --exit-code 0 --format json --output /evidence/dapper.raw.json "$DAPPER_IMAGE"
run_scan image --cache-dir /tmp/trivy-cache --skip-db-update --offline-scan \
    --scanners vuln,secret --list-all-pkgs --vex /workspace/security/runtime-openvex.json \
    --exit-code 0 --format json --output /evidence/dapper.effective.json "$DAPPER_IMAGE"
run_scan image --cache-dir /tmp/trivy-cache --skip-db-update --offline-scan \
    --format cyclonedx --output /evidence/dapper.cdx.json "$DAPPER_IMAGE"
for binary in docker docker-buildx trivy; do
    jq -e --arg binary "$binary" 'any(.Results[]?;
      .Type == "gobinary" and (.Target | split("/") | last) == $binary
      and (.Packages | length > 0))' dist/dapper.raw.json >/dev/null
done

# Preserve the existing, separately evidenced build-only kernel-header boundary.
# A fixed version, wrong package/version or unexpected status is never excluded.
jq '[.Results[]?.Vulnerabilities[]? | select(.PkgName == "linux-libc-dev"
    and .InstalledVersion == "7.0.0-38.38" and (.FixedVersion // "") == ""
    and .Status == "affected")]' dist/dapper.raw.json > dist/dapper.accepted-build-header-findings.json
python3 .github/scripts/enforce-runtime-vulnerabilities.py --phase builder --repository "$GITHUB_REPOSITORY"
docker run --rm --entrypoint /usr/bin/cat "$DAPPER_IMAGE" \
    /licenses/DAPPER-UBUNTU-APT-PACKAGES.tsv > dist/dapper-ubuntu-apt-packages.tsv
cp security/runtime-openvex.json dist/runtime-openvex.json
cp package/ubuntu-apt.lock dist/ubuntu-apt.lock
cp package/ubuntu-security-packages.lock dist/ubuntu-security-packages.lock
printf '%s\n' "$GITHUB_SHA" > dist/source-revision.txt
sha256sum dist/*.json dist/*.tar dist/*.tsv dist/*.lock dist/*-PACKAGES.txt | LC_ALL=C sort -k2 > dist/SHA256SUMS
printf 'KUBECTL_SERVICE_FULL_SEVERITY_GATE_OK fixed_available=0 unregistered=0 secrets=0\n'

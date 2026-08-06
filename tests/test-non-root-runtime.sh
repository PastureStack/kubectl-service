#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
temporary_directory=$(mktemp -d)
cleanup() {
    rm -rf -- "$temporary_directory"
}
trap cleanup EXIT

system_bundle="$temporary_directory/system-ca.pem"
platform_ca="$temporary_directory/platform-ca.pem"
home_directory="$temporary_directory/home"
output_bundle="$home_directory/.local/share/ca-certificates/ca-bundle.crt"
mkdir -p "$home_directory"

cat > "$system_bundle" <<'EOF_SYSTEM'
-----BEGIN CERTIFICATE-----
SYSTEM-CA-TEST-FIXTURE
-----END CERTIFICATE-----
EOF_SYSTEM
cat > "$platform_ca" <<'EOF_PLATFORM'
-----BEGIN CERTIFICATE-----
PLATFORM-CA-TEST-FIXTURE
-----END CERTIFICATE-----
EOF_PLATFORM

HOME="$home_directory" \
SYSTEM_CA_BUNDLE="$system_bundle" \
PLATFORM_CA_ROOT="$platform_ca" \
SSL_CERT_FILE="$output_bundle" \
bash "$repo_root/package/update-platform-ca" "$output_bundle"

test "$(stat -c '%a' "$output_bundle")" = 600
grep -Fqx 'SYSTEM-CA-TEST-FIXTURE' "$output_bundle"
grep -Fqx 'PLATFORM-CA-TEST-FIXTURE' "$output_bundle"

system_only_bundle="$home_directory/.local/share/ca-certificates/system-only.crt"
HOME="$home_directory" \
SYSTEM_CA_BUNDLE="$system_bundle" \
PLATFORM_CA_ROOT="$temporary_directory/missing-platform-ca.pem" \
SSL_CERT_FILE="$system_only_bundle" \
bash "$repo_root/package/update-platform-ca" "$system_only_bundle"
cmp "$system_bundle" "$system_only_bundle"

invalid_ca="$temporary_directory/invalid-ca.pem"
cat > "$invalid_ca" <<'EOF_INVALID'
PRIVATE KEY TEST FIXTURE MUST BE REJECTED
EOF_INVALID
before_hash=$(sha256sum "$output_bundle" | awk '{print $1}')
if HOME="$home_directory" \
    SYSTEM_CA_BUNDLE="$system_bundle" \
    PLATFORM_CA_ROOT="$invalid_ca" \
    SSL_CERT_FILE="$output_bundle" \
    bash "$repo_root/package/update-platform-ca" "$output_bundle" >/dev/null 2>&1; then
    echo 'A private key in the CA input unexpectedly passed validation' >&2
    exit 1
fi
test "$(sha256sum "$output_bundle" | awk '{print $1}')" = "$before_hash"

test "$(awk '$1 == "USER" { user=$2 } END { print user }' "$repo_root/package/Dockerfile")" = '65534:65534'
test "$(awk '$1 == "USER" { user=$2 } END { print user }' "$repo_root/Dockerfile.dapper")" = '65534:65534'
grep -Fq 'io.pasturestack.runtime.user="65534:65534"' "$repo_root/package/Dockerfile"
grep -Fq 'io.pasturestack.shell.requires-privileged="false"' "$repo_root/package/Dockerfile"
grep -Fq 'ARG TILLER_IMAGE_VERSION=v2.17.1' "$repo_root/package/Dockerfile"
grep -Fq 'ARG HELM_BUILD_VERSION=v2.17.1' "$repo_root/package/Dockerfile"
grep -Fq 'HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256=' "$repo_root/migration/targets.lock.env"
grep -Fq 'TILLER_REBUILT_LINUX_AMD64_SHA256=' "$repo_root/migration/targets.lock.env"
grep -Fq 'ARG HELM_MIGRATION_IMAGE_VERSION=v0.11.1' "$repo_root/package/Dockerfile"
grep -Fq 'TILLER_TAG=${TILLER_TAG:-v2.17.1}' "$repo_root/scripts/package"
grep -Fq 'HELM_MIGRATION_TAG=${HELM_MIGRATION_TAG:-v0.11.1}' "$repo_root/scripts/package"
for active_file in migration/targets.lock.env package/Dockerfile scripts/package scripts/release README.md SECURITY.md COMPATIBILITY.md MIGRATION.md; do
    if grep -Eq 'v[0-9]+\.[0-9]+\.[0-9]+(-pasturestack\.[0-9]+|\+pasturestack([.-][0-9A-Za-z.-]+)?)' "$repo_root/$active_file"; then
        echo "An active artifact version still contains a brand suffix: $active_file" >&2
        exit 1
    fi
done
grep -Fq "test \"\$(docker image inspect \"\${IMAGE}\" --format '{{.Config.User}}')\" = \"65534:65534\"" "$repo_root/scripts/package"
grep -Fq -- '--user "$$(id -u):$$(id -g)"' "$repo_root/Makefile"
grep -Fq -- '--group-add "$$(stat -c '\''%g'\'' /var/run/docker.sock)"' "$repo_root/Makefile"
grep -Fq -- '-e HOME=/tmp/pasturestack-dapper-home-$$(id -u)' "$repo_root/Makefile"
grep -Fq -- '-e GOCACHE=/tmp/go-build-cache-$$(id -u)' "$repo_root/Makefile"
grep -Fq -- '-e GOMAXPROCS=2' "$repo_root/Makefile"
grep -Fq 'DOCKER_BUILDKIT ?= 1' "$repo_root/Makefile"
grep -Fq -- '-e DOCKER_BUILDKIT=$(DOCKER_BUILDKIT)' "$repo_root/Makefile"
grep -Fq 'install -d -m 0700 "${HOME}" "${GOCACHE}" "${XDG_CONFIG_HOME}"' "$repo_root/scripts/entry"
grep -Fq 'FROM docker/buildx-bin:0.36.1@sha256:1f2f6b2be4a2511ada67336e76892f1a588c89746009dd4b21069e4d867465be AS buildx' "$repo_root/Dockerfile.dapper"
grep -Fq 'ARG BUILDX_GIT_COMMIT=1d8dde89b8aba914e05e45366770736fea1fd690' "$repo_root/Dockerfile.dapper"
grep -Fq 'COPY --from=buildx /buildx /usr/libexec/docker/cli-plugins/docker-buildx' "$repo_root/Dockerfile.dapper"
grep -Fq 'docker buildx version | grep -F "github.com/docker/buildx v${BUILDX_VERSION} ${BUILDX_GIT_COMMIT}"' "$repo_root/Dockerfile.dapper"
grep -Fq 'FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS trivy-builder' "$repo_root/Dockerfile.dapper"
for stage in 'helm2-builder' 'tiller-runtime' 'kubectl-extractor' 'helm-migration-tools-runtime'; do
    grep -Fq "FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS ${stage}" "$repo_root/package/Dockerfile"
done
grep -Fqx 'FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03' "$repo_root/package/Dockerfile"
grep -Fq 'ARG TRIVY_GIT_COMMIT=40c73e5d6166dcc0346a1ab4e94499d1572854e4' "$repo_root/Dockerfile.dapper"
grep -Fq 'ARG TRIVY_SOURCE_SHA256=dd35bec9570e1968b7a3d0d9f6504e5ac2f6a87eea0eee8ddcadd44d08940ee7' "$repo_root/Dockerfile.dapper"
grep -Fq 'ARG TRIVY_VERSION=0.73.1' "$repo_root/Dockerfile.dapper"
grep -Fq 'ARG TRIVY_PATCH_SHA256=ae0430b03044dff4b9b2e4115f1e5c156f623356aeb0e1213e4ae0f2c9dd8354' "$repo_root/Dockerfile.dapper"
grep -Fq 'COPY build-tools/trivy-v0.73.0-oras-go-v2.6.2-go-git-v5.19.2.patch /tmp/trivy-security.patch' "$repo_root/Dockerfile.dapper"
echo 'ae0430b03044dff4b9b2e4115f1e5c156f623356aeb0e1213e4ae0f2c9dd8354  build-tools/trivy-v0.73.0-oras-go-v2.6.2-go-git-v5.19.2.patch' | \
    (cd "$repo_root" && sha256sum -c -)
grep -Fq 'GOMAXPROCS=2' "$repo_root/Dockerfile.dapper"
grep -Fq 'GOEXPERIMENT=jsonv2 CGO_ENABLED=0 go build -p=2 -mod=readonly -trimpath -buildvcs=false' "$repo_root/Dockerfile.dapper"
grep -Fq 'TRIVY-BUILDINFO.txt' "$repo_root/Dockerfile.dapper"
grep -Fq '$1 == "dep" && $2 == "github.com/go-git/go-git/v5" && $3 == "v5.19.2"' "$repo_root/Dockerfile.dapper"
grep -Fq '$1 == "dep" && $2 == "oras.land/oras-go/v2" && $3 == "v2.6.2"' "$repo_root/Dockerfile.dapper"
grep -Fq 'COPY --from=trivy-builder /out/trivy /usr/local/bin/trivy' "$repo_root/Dockerfile.dapper"
grep -Fq 'COPY --from=trivy-builder /out/TRIVY-BUILDINFO.txt /licenses/TRIVY-BUILDINFO.txt' "$repo_root/Dockerfile.dapper"
grep -Fq 'test "$(trivy --version)" = "Version: ${TRIVY_VERSION}"' "$repo_root/Dockerfile.dapper"
if grep -Fq 'FROM aquasec/trivy:' "$repo_root/Dockerfile.dapper"; then
    echo 'The Dapper image still trusts the vulnerable published Trivy binary' >&2
    exit 1
fi
echo '7d1151fa1cc9f9a383815454c3c3322c05c499cd76982f8bd4289547869616bd  build-tools/buildx-v0.36.1-docker-module-packages.txt' | \
    (cd "$repo_root" && sha256sum -c -)
for vulnerability in CVE-2026-34040 CVE-2026-41567 CVE-2026-42306; do
    grep -Fq "\"name\": \"${vulnerability}\"" "$repo_root/security/openvex.json"
done
grep -Fq 'pkg:golang/github.com/docker/docker@v28.5.2%2Bincompatible' "$repo_root/security/openvex.json"
test "$(grep -c 'pkg:deb/ubuntu/linux-libc-dev@7.0.0-29.29?arch=amd64&distro=ubuntu-26.04' "$repo_root/security/openvex.json")" -eq 46
grep -Fq 'CVE-2026-53215' "$repo_root/security/openvex.json"
grep -Fq 'CVE-2026-53260' "$repo_root/security/openvex.json"
grep -Fq 'DAPPER_TRIVY_CACHE ?= $(HOME)/.cache/trivy' "$repo_root/Makefile"
grep -Fq -- '-v "$(DAPPER_TRIVY_CACHE):/tmp/trivy-cache"' "$repo_root/Makefile"
grep -Fq -- '-e TRIVY_CACHE_DIR=/tmp/trivy-cache' "$repo_root/Makefile"
grep -Fq 'ci package release: dapper-trivy-cache' "$repo_root/Makefile"
for required in \
    'gcc="${UBUNTU_APT_GCC_VERSION}" \' \
    'libc6-dev="${UBUNTU_APT_LIBC6_DEV_VERSION}" \' ; do
    if ! grep -Fq -- "$required" "$repo_root/Dockerfile.dapper"; then
        echo "The Dapper race-test toolchain is not exactly version-locked: $required" >&2
        exit 1
    fi
done
grep -Fq 'go test -race -cover -tags=test' "$repo_root/scripts/test"

if grep -Eq '(^|[[:space:]])chown([[:space:]]|$)' "$repo_root/scripts/entry"; then
    echo 'The build entrypoint still rewrites mounted source ownership as root' >&2
    exit 1
fi

if grep -Eq '(^|[[:space:]])(chown|su)([[:space:]]|$)' "$repo_root/package/kubectl-service.sh"; then
    echo 'The service entrypoint still requires a root-only ownership or user switch operation' >&2
    exit 1
fi
if grep -Eq '(^|[[:space:]])(unshare|mount|chown|su)([[:space:]]|$)' \
    "$repo_root/package/kubectl-shell.sh" "$repo_root/package/shell-setup.sh"; then
    echo 'The shell session path still requires a privileged namespace or ownership operation' >&2
    exit 1
fi
grep -Fq 'mktemp -d "${session_root}/pasturestack-kubectl-shell.XXXXXX"' "$repo_root/package/shell-setup.sh"
grep -Fq 'KUBECONFIG="${session_home}/.kube/config"' "$repo_root/package/shell-setup.sh"

shell_session_root="$temporary_directory/sessions"
mkdir -p "$shell_session_root"
if KUBECTL_SHELL_TOKEN=test-token \
    KUBECTL_SHELL_SESSION_ROOT="$shell_session_root" \
    KUBERNETES_URL='https://user:password@kubernetes.example.invalid:6443/' \
    bash "$repo_root/package/shell-setup.sh" >/dev/null 2>&1; then
    echo 'A shell Kubernetes URL containing embedded credentials unexpectedly passed validation' >&2
    exit 1
fi
test "$(find "$shell_session_root" -mindepth 1 -maxdepth 1 -type d -name 'pasturestack-kubectl-shell.*' -print | wc -l)" -eq 0

echo 'KUBECTL_SERVICE_NON_ROOT_RUNTIME_TEST_OK user=65534:65534 privileged=false'

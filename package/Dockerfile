FROM ubuntu:26.04 AS helm2-builder

ARG UBUNTU_MIRROR=http://archive.ubuntu.com/ubuntu
ARG GO_VERSION=1.26.5
ARG GO_SHA256_amd64=5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053
ARG HELM_VERSION=v2.17.0
ARG HELM_GIT_COMMIT=a690bad98af45b015bd3da1a41f6218b1a451dbe

ENV DEBIAN_FRONTEND=noninteractive \
    GOPATH=/go \
    GO111MODULE=on \
    GOCACHE=/tmp/go-build-cache \
    GOMODCACHE=/tmp/go-mod-cache \
    GOTELEMETRY=off \
    XDG_CONFIG_HOME=/tmp/go-config \
    GIT_CONFIG_GLOBAL=/tmp/gitconfig \
    PATH=/go/bin:/usr/local/go/bin:${PATH}

RUN set -eux; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e "s|http://security.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" \
        -e "s|http://archive.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        gcc \
        git \
        libc6-dev \
        tar; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/go.tgz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"; \
    echo "${GO_SHA256_amd64}  /tmp/go.tgz" | sha256sum -c -; \
    tar -C /usr/local -xzf /tmp/go.tgz; \
    rm -f /tmp/go.tgz; \
    mkdir -p /go/bin /go/src /out; \
    git config --system --add safe.directory '*'; \
    go version

ARG HELM_PATCH_SHA256=3bdf5475025db4d76afad26272f4848b2604c6e0e547b3aaf8243b48e16a9735

COPY helm-v2.17.0.go.mod /tmp/helm-lock/test.go.mod
COPY helm-v2.17.0.go.sum /tmp/helm-lock/test.go.sum
COPY helm-v2.17.0-client.go.mod /tmp/helm-lock/client.go.mod
COPY helm-v2.17.0-client.go.sum /tmp/helm-lock/client.go.sum
COPY helm-v2.17.0-pasturestack.patch /tmp/helm-lock/helm-v2.17.0-pasturestack.patch

RUN set -eux; \
    mkdir -p /go/src/k8s.io; \
    git clone --depth 1 --branch "${HELM_VERSION}" https://github.com/helm/helm /go/src/k8s.io/helm; \
    cd /go/src/k8s.io/helm; \
    test "$(git rev-parse HEAD)" = "${HELM_GIT_COMMIT}"; \
    echo "${HELM_PATCH_SHA256}  /tmp/helm-lock/helm-v2.17.0-pasturestack.patch" | sha256sum -c -; \
    git apply --check /tmp/helm-lock/helm-v2.17.0-pasturestack.patch; \
    git apply /tmp/helm-lock/helm-v2.17.0-pasturestack.patch; \
    git diff --check; \
    cp /tmp/helm-lock/test.go.mod ./go.mod; \
    cp /tmp/helm-lock/test.go.sum ./go.sum; \
    go mod download; \
    go mod verify; \
    test -z "$(gofmt -l \
        pkg/kube/client.go \
        pkg/kube/wait.go \
        pkg/kube/pasturestack_humanreadable_flags.go \
        pkg/kube/pasturestack_table_printer.go \
        pkg/pasturestack/printers)"; \
    go test ./pkg/kube ./pkg/pasturestack/printers; \
    CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=pasturestack-go${GO_VERSION}" \
        -o /out/tiller ./cmd/tiller; \
    cp LICENSE /out/HELM-LICENSE; \
    ! go version -m /out/tiller | grep -Fq 'github.com/dgrijalva/jwt-go'; \
    go version -m /out/tiller | grep -F 'github.com/prometheus/client_golang' | grep -Fq 'v1.24.1'; \
    /out/tiller --version; \
    cp /tmp/helm-lock/client.go.mod ./go.mod; \
    cp /tmp/helm-lock/client.go.sum ./go.sum; \
    go mod download; \
    go mod verify; \
    go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=pasturestack-go${GO_VERSION}" \
        -o /out/helm ./cmd/helm; \
    ! go version -m /out/helm | grep -Fq 'k8s.io/kubernetes'; \
    ! go version -m /out/helm | grep -Fq 'github.com/dgrijalva/jwt-go'; \
    ! go version -m /out/helm | grep -Fq 'github.com/prometheus/client_golang'; \
    go version -m /out/helm | grep -Fq 'github.com/moby/term'; \
    go version -m /out/helm | grep -Fq 'google.golang.org/grpc'; \
    /out/helm version --client; \
    HOME=/tmp/helm-home /out/helm init -c; \
    rm -rf \
        /go/src/k8s.io/helm \
        /tmp/go-build-cache \
        /tmp/go-mod-cache \
        /tmp/helm-home \
        /root/.cache

FROM ubuntu:26.04 AS tiller-runtime

ARG UBUNTU_MIRROR=http://archive.ubuntu.com/ubuntu
ARG TILLER_IMAGE_VERSION=v2.17.0-pasturestack.2
ARG SOURCE_REVISION=unknown

ENV DEBIAN_FRONTEND=noninteractive \
    HOME=/home/pasturestack-tiller \
    TILLER_NAMESPACE=kube-system

RUN set -eux; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e "s|http://security.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" \
        -e "s|http://archive.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" {} +; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble; \
    groupadd --gid 10001 pasturestack-tiller; \
    useradd --uid 10001 --gid 10001 --home-dir /home/pasturestack-tiller --shell /usr/sbin/nologin pasturestack-tiller; \
    install -d -o 10001 -g 10001 /home/pasturestack-tiller /licenses

COPY --from=helm2-builder /out/tiller /usr/local/bin/tiller
COPY --from=helm2-builder /out/HELM-LICENSE /licenses/HELM-LICENSE
COPY MIGRATION.md targets.lock.env kubernetes-upgrade-path.tsv /licenses/

RUN chmod 0555 /usr/local/bin/tiller && chmod 0444 /licenses/HELM-LICENSE

USER 10001:10001

EXPOSE 44134 44135

LABEL org.opencontainers.image.title="pasturestack-tiller-compat" \
      org.opencontainers.image.description="PastureStack Helm 2.17 Tiller compatibility runtime" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/kubectl-service" \
      org.opencontainers.image.version="${TILLER_IMAGE_VERSION}" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      io.pasturestack.compatibility="helm2"

STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/tiller"]

FROM ubuntu:26.04 AS kubectl-extractor

ARG UBUNTU_MIRROR=http://archive.ubuntu.com/ubuntu
ARG KUBERNETES_VERSION=v1.12.10-rancher1
ARG KUBERNETES_SERVER_SHA256=f3df06f9bf2e4c4e8b5aba16d44197e26672aa06e8c24ab9f6ab4eaf64ab068e

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e "s|http://security.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" \
        -e "s|http://archive.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        tar; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/kubernetes-server.tgz \
        "https://github.com/rancher/kubernetes/releases/download/${KUBERNETES_VERSION}/kubernetes-server-linux-amd64.tar.gz"; \
    echo "${KUBERNETES_SERVER_SHA256}  /tmp/kubernetes-server.tgz" | sha256sum -c -; \
    mkdir -p /out; \
    tar -xzf /tmp/kubernetes-server.tgz -C /out --strip-components=3 \
        kubernetes/server/bin/kubectl; \
    chmod +x /out/kubectl; \
    /out/kubectl version --client=true --short=true; \
    rm -f /tmp/kubernetes-server.tgz

FROM helm2-builder AS helm3-bridge-builder

ARG GO_VERSION=1.26.5
ARG HELM3_BRIDGE_VERSION=v3.21.3
ARG HELM3_BRIDGE_GIT_COMMIT=1ad6e68924fdf6fb0c7dcef8e9e1dfc0f36eaed6
ARG HELM3_BRIDGE_SOURCE_SHA256=526f2645c652a3ba2b9b6f38f8d2bab47be4626d4850dafb4311e5f50f6252ef
ARG HELM3_BRIDGE_LICENSE_SHA256=881467e52efeb1807406134b0ec04b1c6f52dc9dc49382713ef9ac4e0cae42f8
ARG HELM3_BRIDGE_APPLIED_PATCH_SHA256=7ea5776330f4af37cc00bb7ea4c2669215fcd5c6b450ca632dc1213c591fe5b2
ARG HELM3_BRIDGE_ORAS_GO_VERSION=v2.6.2
ARG HELM3_BRIDGE_X_TEXT_VERSION=v0.39.0
ARG HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=7ff975f19c1af44f1a0cd7a2738e363752acd713be77e7c0b02b653946e5c833

ENV GOMAXPROCS=2

COPY helm-v3.21.3-security.patch /tmp/helm3-security.patch

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/helm3-source.tgz \
        "https://github.com/helm/helm/archive/${HELM3_BRIDGE_GIT_COMMIT}.tar.gz"; \
    echo "${HELM3_BRIDGE_SOURCE_SHA256}  /tmp/helm3-source.tgz" | sha256sum -c -; \
    echo "${HELM3_BRIDGE_APPLIED_PATCH_SHA256}  /tmp/helm3-security.patch" | sha256sum -c -; \
    test "$(grep -c '^--- a/' /tmp/helm3-security.patch)" -eq 2; \
    test "$(grep -c '^+++ b/' /tmp/helm3-security.patch)" -eq 2; \
    grep -Fqx -- '--- a/go.mod' /tmp/helm3-security.patch; \
    grep -Fqx -- '+++ b/go.mod' /tmp/helm3-security.patch; \
    grep -Fqx -- '--- a/go.sum' /tmp/helm3-security.patch; \
    grep -Fqx -- '+++ b/go.sum' /tmp/helm3-security.patch; \
    grep -Fq 'oras.land/oras-go/v2 v2.6.1' /tmp/helm3-security.patch; \
    grep -Fq "oras.land/oras-go/v2 ${HELM3_BRIDGE_ORAS_GO_VERSION}" /tmp/helm3-security.patch; \
    grep -Fq 'golang.org/x/text v0.38.0' /tmp/helm3-security.patch; \
    grep -Fq "golang.org/x/text ${HELM3_BRIDGE_X_TEXT_VERSION}" /tmp/helm3-security.patch; \
    mkdir -p /tmp/helm3-source; \
    tar -xzf /tmp/helm3-source.tgz -C /tmp/helm3-source --strip-components=1; \
    cd /tmp/helm3-source; \
    test "$(go env GOVERSION)" = "go${GO_VERSION}"; \
    echo "${HELM3_BRIDGE_LICENSE_SHA256}  LICENSE" | sha256sum -c -; \
    git apply --check --whitespace=error-all /tmp/helm3-security.patch; \
    git apply --whitespace=error-all /tmp/helm3-security.patch; \
    test "$(go list -m -f '{{.GoVersion}}')" = '1.26.0'; \
    test "$(awk '$1 == "oras.land/oras-go/v2" {print $2}' go.mod)" = "${HELM3_BRIDGE_ORAS_GO_VERSION}"; \
    test "$(awk '$1 == "golang.org/x/text" {print $2}' go.mod)" = "${HELM3_BRIDGE_X_TEXT_VERSION}"; \
    go mod download all; \
    go mod verify; \
    install -d -m 0755 -o nobody -g nogroup \
        /tmp/helm3-test-cache \
        /tmp/helm3-test-gopath; \
    su -s /bin/bash nobody -c \
        'cd /tmp/helm3-source && env GOPATH=/tmp/helm3-test-gopath GOCACHE=/tmp/helm3-test-cache GOMODCACHE=/tmp/go-mod-cache GOPROXY=off GOSUMDB=off GOTELEMETRY=off /usr/local/go/bin/go test -p=2 -mod=readonly -count=1 ./internal/version ./pkg/action ./pkg/storage/...'; \
    rm -rf /tmp/helm3-test-cache /tmp/helm3-test-gopath; \
    env GOCACHE=/tmp/helm3-build-cache-a GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X helm.sh/helm/v3/internal/version.version=${HELM3_BRIDGE_VERSION} -X helm.sh/helm/v3/internal/version.metadata=pasturestack.security.1 -X helm.sh/helm/v3/internal/version.gitCommit=${HELM3_BRIDGE_GIT_COMMIT} -X helm.sh/helm/v3/internal/version.gitTreeState=clean" \
        -o /out/helm3-first ./cmd/helm; \
    rm -rf /tmp/helm3-build-cache-a; \
    env GOCACHE=/tmp/helm3-build-cache-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X helm.sh/helm/v3/internal/version.version=${HELM3_BRIDGE_VERSION} -X helm.sh/helm/v3/internal/version.metadata=pasturestack.security.1 -X helm.sh/helm/v3/internal/version.gitCommit=${HELM3_BRIDGE_GIT_COMMIT} -X helm.sh/helm/v3/internal/version.gitTreeState=clean" \
        -o /out/helm3-second ./cmd/helm; \
    rm -rf /tmp/helm3-build-cache-b; \
    cmp /out/helm3-first /out/helm3-second; \
    echo "${HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256}  /out/helm3-first" | sha256sum -c -; \
    mv /out/helm3-first /out/helm3; \
    rm -f /out/helm3-second; \
    go version -m /out/helm3 > /out/HELM3-MODULES.txt; \
    cp LICENSE /out/HELM3-LICENSE; \
    test "$(awk '$1 == "dep" && $2 == "oras.land/oras-go/v2" {print $3}' /out/HELM3-MODULES.txt)" = "${HELM3_BRIDGE_ORAS_GO_VERSION}"; \
    test "$(awk '$1 == "dep" && $2 == "golang.org/x/text" {print $3}' /out/HELM3-MODULES.txt)" = "${HELM3_BRIDGE_X_TEXT_VERSION}"; \
    test "$(/out/helm3 version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }}')" = "${HELM3_BRIDGE_VERSION}+pasturestack.security.1 ${HELM3_BRIDGE_GIT_COMMIT} clean"; \
    chmod 0555 /out/helm3; \
    chmod 0444 /out/HELM3-LICENSE /out/HELM3-MODULES.txt; \
    rm -rf \
        /tmp/helm3-source \
        /tmp/helm3-source.tgz \
        /tmp/helm3-security.patch \
        /tmp/helm3-test-cache \
        /tmp/helm3-test-gopath \
        /tmp/helm3-build-cache-a \
        /tmp/helm3-build-cache-b \
        /tmp/go-build-cache \
        /tmp/go-mod-cache \
        /root/.cache

FROM helm3-bridge-builder AS helm2to3-builder

ARG GO_VERSION=1.26.5
ARG HELM2TO3_VERSION=v0.11.0
ARG HELM2TO3_GIT_COMMIT=98df8bee4094de2e81c0507108867d8139d1306b
ARG HELM2TO3_SOURCE_SHA256=135aab4d90991fe088f3dc1ecf0c1419400d1035068538ef25337c2832080a08
ARG HELM2TO3_LICENSE_SHA256=c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4
ARG HELM2TO3_APPLIED_PATCH_SHA256=d0e5cc51d9ac3e97df4df493e01971f7aa1d9010951aac3edeefc526cc7e3d2d
ARG HELM2TO3_HELM3_VERSION=v3.21.3
ARG HELM2TO3_ORAS_GO_VERSION=v2.6.2
ARG HELM2TO3_X_TEXT_VERSION=v0.39.0
ARG HELM2TO3_REBUILT_LINUX_AMD64_SHA256=23db138e69dc25c15de175c5cbf0b2ae112c079b7d92124da86b8757d4c2c72c

COPY helm-2to3-v0.11.0-security.patch /tmp/helm2to3-security.patch

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/helm2to3-source.tgz \
        "https://github.com/helm/helm-2to3/archive/${HELM2TO3_GIT_COMMIT}.tar.gz"; \
    echo "${HELM2TO3_SOURCE_SHA256}  /tmp/helm2to3-source.tgz" | sha256sum -c -; \
    echo "${HELM2TO3_APPLIED_PATCH_SHA256}  /tmp/helm2to3-security.patch" | sha256sum -c -; \
    test "$(grep -c '^--- a/' /tmp/helm2to3-security.patch)" -eq 2; \
    test "$(grep -c '^+++ b/' /tmp/helm2to3-security.patch)" -eq 2; \
    grep -Fqx -- '--- a/go.mod' /tmp/helm2to3-security.patch; \
    grep -Fqx -- '+++ b/go.mod' /tmp/helm2to3-security.patch; \
    grep -Fqx -- '--- a/go.sum' /tmp/helm2to3-security.patch; \
    grep -Fqx -- '+++ b/go.sum' /tmp/helm2to3-security.patch; \
    grep -Fq 'helm.sh/helm/v3 v3.15.3' /tmp/helm2to3-security.patch; \
    grep -Fq "helm.sh/helm/v3 ${HELM2TO3_HELM3_VERSION}" /tmp/helm2to3-security.patch; \
    grep -Fq 'oras.land/oras-go v1.2.5' /tmp/helm2to3-security.patch; \
    grep -Fq "oras.land/oras-go/v2 ${HELM2TO3_ORAS_GO_VERSION}" /tmp/helm2to3-security.patch; \
    grep -Fq 'golang.org/x/text v0.14.0' /tmp/helm2to3-security.patch; \
    grep -Fq "golang.org/x/text ${HELM2TO3_X_TEXT_VERSION}" /tmp/helm2to3-security.patch; \
    mkdir -p /tmp/helm2to3-source; \
    tar -xzf /tmp/helm2to3-source.tgz -C /tmp/helm2to3-source --strip-components=1; \
    cd /tmp/helm2to3-source; \
    test "$(go env GOVERSION)" = "go${GO_VERSION}"; \
    echo "${HELM2TO3_LICENSE_SHA256}  LICENSE" | sha256sum -c -; \
    git apply --check --whitespace=error-all /tmp/helm2to3-security.patch; \
    git apply --whitespace=error-all /tmp/helm2to3-security.patch; \
    test "$(go list -m -f '{{.GoVersion}}')" = '1.26.0'; \
    grep -Fqx 'toolchain go1.26.5' go.mod; \
    test "$(awk '$1 == "helm.sh/helm/v3" {print $2}' go.mod)" = "${HELM2TO3_HELM3_VERSION}"; \
    test "$(awk '$1 == "oras.land/oras-go/v2" {print $2}' go.mod)" = "${HELM2TO3_ORAS_GO_VERSION}"; \
    test "$(awk '$1 == "golang.org/x/text" {print $2}' go.mod)" = "${HELM2TO3_X_TEXT_VERSION}"; \
    go mod download all; \
    go mod verify; \
    install -d -m 0755 -o nobody -g nogroup \
        /tmp/helm2to3-test-cache \
        /tmp/helm2to3-test-gopath; \
    su -s /bin/bash nobody -c \
        'cd /tmp/helm2to3-source && env GOPATH=/tmp/helm2to3-test-gopath GOCACHE=/tmp/helm2to3-test-cache GOMODCACHE=/tmp/go-mod-cache GOPROXY=off GOSUMDB=off GOTELEMETRY=off /usr/local/go/bin/go test -p=2 -mod=readonly -count=1 ./...'; \
    rm -rf /tmp/helm2to3-test-cache /tmp/helm2to3-test-gopath; \
    env GOCACHE=/tmp/helm2to3-build-cache-a GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X main.version=${HELM2TO3_VERSION}+pasturestack.security.1" \
        -o /out/helm2to3-first ./main.go; \
    rm -rf /tmp/helm2to3-build-cache-a; \
    env GOCACHE=/tmp/helm2to3-build-cache-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X main.version=${HELM2TO3_VERSION}+pasturestack.security.1" \
        -o /out/helm2to3-second ./main.go; \
    rm -rf /tmp/helm2to3-build-cache-b; \
    cmp /out/helm2to3-first /out/helm2to3-second; \
    echo "${HELM2TO3_REBUILT_LINUX_AMD64_SHA256}  /out/helm2to3-first" | sha256sum -c -; \
    mv /out/helm2to3-first /out/helm2to3; \
    rm -f /out/helm2to3-second; \
    go version -m /out/helm2to3 > /out/HELM2TO3-MODULES.txt; \
    cp LICENSE /out/HELM2TO3-LICENSE; \
    test "$(awk '$1 == "dep" && $2 == "helm.sh/helm/v3" {print $3}' /out/HELM2TO3-MODULES.txt)" = "${HELM2TO3_HELM3_VERSION}"; \
    test "$(awk '$1 == "dep" && $2 == "oras.land/oras-go/v2" {print $3}' /out/HELM2TO3-MODULES.txt)" = "${HELM2TO3_ORAS_GO_VERSION}"; \
    test "$(awk '$1 == "dep" && $2 == "golang.org/x/text" {print $3}' /out/HELM2TO3-MODULES.txt)" = "${HELM2TO3_X_TEXT_VERSION}"; \
    /out/helm2to3 --help | grep -Fq 'Migrate and Cleanup Helm v2 configuration and releases in-place to Helm v3'; \
    chmod 0555 /out/helm2to3; \
    chmod 0444 /out/HELM2TO3-LICENSE /out/HELM2TO3-MODULES.txt; \
    rm -rf \
        /tmp/helm2to3-source \
        /tmp/helm2to3-source.tgz \
        /tmp/helm2to3-security.patch \
        /tmp/helm2to3-test-cache \
        /tmp/helm2to3-test-gopath \
        /tmp/helm2to3-build-cache-a \
        /tmp/helm2to3-build-cache-b \
        /tmp/go-build-cache \
        /tmp/go-mod-cache \
        /root/.cache

FROM ubuntu:26.04 AS helm-migration-tools-runtime

ARG HELM_MIGRATION_IMAGE_VERSION=v0.11.0-pasturestack.1
ARG SOURCE_REVISION=unknown
ARG HELM3_BRIDGE_VERSION=v3.21.3
ARG HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=7ff975f19c1af44f1a0cd7a2738e363752acd713be77e7c0b02b653946e5c833
ARG HELM2TO3_VERSION=v0.11.0
ARG HELM2TO3_REBUILT_LINUX_AMD64_SHA256=23db138e69dc25c15de175c5cbf0b2ae112c079b7d92124da86b8757d4c2c72c

RUN install -d -m 0755 /licenses

COPY --from=helm3-bridge-builder /out/helm3 /usr/local/bin/helm3
COPY --from=helm3-bridge-builder /out/HELM3-LICENSE /licenses/HELM3-LICENSE
COPY --from=helm3-bridge-builder /out/HELM3-MODULES.txt /licenses/HELM3-MODULES.txt
COPY --from=helm2to3-builder /out/helm2to3 /usr/local/bin/helm2to3
COPY --from=helm2to3-builder /out/HELM2TO3-LICENSE /licenses/HELM2TO3-LICENSE
COPY --from=helm2to3-builder /out/HELM2TO3-MODULES.txt /licenses/HELM2TO3-MODULES.txt
COPY SECURITY.md MIGRATION.md targets.lock.env kubernetes-upgrade-path.tsv /licenses/

RUN set -eux; \
    rm -f /usr/bin/pebble; \
    echo "${HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256}  /usr/local/bin/helm3" | sha256sum -c -; \
    echo "${HELM2TO3_REBUILT_LINUX_AMD64_SHA256}  /usr/local/bin/helm2to3" | sha256sum -c -; \
    chmod 0555 /usr/local/bin/helm3 /usr/local/bin/helm2to3; \
    chmod 0444 /licenses/*

USER 10001:10001

LABEL org.opencontainers.image.title="pasturestack-helm-migration-tools" \
      org.opencontainers.image.description="Explicit-use, source-rebuilt Helm 2 to Helm 3 migration tools" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/kubectl-service" \
      org.opencontainers.image.version="${HELM_MIGRATION_IMAGE_VERSION}" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      io.pasturestack.helm3.version="${HELM3_BRIDGE_VERSION}" \
      io.pasturestack.helm2to3.version="${HELM2TO3_VERSION}" \
      io.pasturestack.execution="explicit-only" \
      io.pasturestack.upstream-binaries="prohibited"

ENTRYPOINT ["/usr/bin/false"]

FROM helm2-builder AS helm4-builder

ARG GO_VERSION=1.26.5
ARG HELM4_TARGET_VERSION=v4.2.3
ARG HELM4_TARGET_GIT_COMMIT=43e8b7feece8beb0fcba47059ec9b522fd929a64
ARG HELM4_TARGET_SOURCE_SHA256=9ca555fa3777082fd646c95abf4ee17fc31e25ba97bad76da9de1df2332be168
ARG HELM4_TARGET_LICENSE_SHA256=881467e52efeb1807406134b0ec04b1c6f52dc9dc49382713ef9ac4e0cae42f8
ARG HELM4_SECURITY_PATCH_COMMIT=84e63e5c38913594e476b15609fb6c7ab2d60467
ARG HELM4_SECURITY_PATCH_SHA256=969f510d0f7a6c0b2afad1c1d9b925fd78838923cc1e6b03e01135822230f515
ARG HELM4_APPLIED_PATCH_SHA256=085bde46aa759a7df8a5a7506e5d21f25ccb5ff5b4a47078c21651eaa4673b6e
ARG HELM4_ORAS_GO_VERSION=v2.6.2

COPY helm-v4.2.3-oras-go-v2.6.2.patch /tmp/helm4-applied.patch

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/helm4-source.tgz "https://github.com/helm/helm/archive/refs/tags/${HELM4_TARGET_VERSION}.tar.gz"; \
    echo "${HELM4_TARGET_SOURCE_SHA256}  /tmp/helm4-source.tgz" | sha256sum -c -; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/helm4-upstream.patch "https://github.com/helm/helm/commit/${HELM4_SECURITY_PATCH_COMMIT}.patch"; \
    echo "${HELM4_SECURITY_PATCH_SHA256}  /tmp/helm4-upstream.patch" | sha256sum -c -; \
    echo "${HELM4_APPLIED_PATCH_SHA256}  /tmp/helm4-applied.patch" | sha256sum -c -; \
    for patch_file in /tmp/helm4-upstream.patch /tmp/helm4-applied.patch; do \
        test "$(grep -c '^diff --git ' "${patch_file}")" -eq 2; \
        grep -Fqx 'diff --git a/go.mod b/go.mod' "${patch_file}"; \
        grep -Fqx 'diff --git a/go.sum b/go.sum' "${patch_file}"; \
        grep -Fq 'oras.land/oras-go/v2 v2.6.1' "${patch_file}"; \
        grep -Fq 'oras.land/oras-go/v2 v2.6.2' "${patch_file}"; \
    done; \
    mkdir -p /tmp/helm4-source; \
    tar -xzf /tmp/helm4-source.tgz -C /tmp/helm4-source --strip-components=1; \
    cd /tmp/helm4-source; \
    test "$(go env GOVERSION)" = "go${GO_VERSION}"; \
    git apply --check --whitespace=error-all /tmp/helm4-applied.patch; \
    git apply --whitespace=error-all /tmp/helm4-applied.patch; \
    test "$(go list -m -f '{{.GoVersion}}')" = '1.26.0'; \
    test "$(awk '$1 == "oras.land/oras-go/v2" {print $2}' go.mod)" = "${HELM4_ORAS_GO_VERSION}"; \
    go mod download; \
    go mod verify; \
    install -d -m 0755 -o nobody -g nogroup \
        /tmp/helm4-test-cache \
        /tmp/helm4-test-gopath \
        /tmp/helm4-test-home; \
    su -s /bin/bash nobody -c \
        'cd /tmp/helm4-source && env HOME=/tmp/helm4-test-home GOPATH=/tmp/helm4-test-gopath GOCACHE=/tmp/helm4-test-cache GOMODCACHE=/tmp/go-mod-cache GOTELEMETRY=off /usr/local/go/bin/go test -mod=readonly oras.land/oras-go/v2/content/file'; \
    go test -mod=readonly ./pkg/registry; \
    CGO_ENABLED=0 go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X helm.sh/helm/v4/internal/version.version=${HELM4_TARGET_VERSION} -X helm.sh/helm/v4/internal/version.metadata=pasturestack.oras-go-${HELM4_ORAS_GO_VERSION#v} -X helm.sh/helm/v4/internal/version.gitCommit=${HELM4_TARGET_GIT_COMMIT} -X helm.sh/helm/v4/internal/version.gitTreeState=clean" \
        -o /out/helm4 ./cmd/helm; \
    go version -m /out/helm4 > /out/HELM4-MODULES.txt; \
    cp LICENSE /out/HELM4-LICENSE; \
    echo "${HELM4_TARGET_LICENSE_SHA256}  /out/HELM4-LICENSE" | sha256sum -c -; \
    test "$(awk '$1 == "dep" && $2 == "oras.land/oras-go/v2" {print $3}' /out/HELM4-MODULES.txt)" = "${HELM4_ORAS_GO_VERSION}"; \
    ! grep -Fq 'oras.land/oras-go/v2 v2.6.1' /out/HELM4-MODULES.txt; \
    test "$(/out/helm4 version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }}')" = "${HELM4_TARGET_VERSION}+pasturestack.oras-go-${HELM4_ORAS_GO_VERSION#v} ${HELM4_TARGET_GIT_COMMIT} clean"; \
    chmod 0555 /out/helm4; \
    chmod 0444 /out/HELM4-LICENSE /out/HELM4-MODULES.txt; \
    rm -rf \
        /tmp/helm4-source \
        /tmp/helm4-source.tgz \
        /tmp/helm4-upstream.patch \
        /tmp/helm4-applied.patch \
        /tmp/helm4-test-cache \
        /tmp/helm4-test-gopath \
        /tmp/helm4-test-home \
        /tmp/go-build-cache \
        /tmp/go-mod-cache \
        /root/.cache

FROM ubuntu:26.04

ARG UBUNTU_MIRROR=http://archive.ubuntu.com/ubuntu
ARG IMAGE_VERSION=dev
ARG SOURCE_REVISION=unknown
ARG HELM4_TARGET_VERSION=v4.2.3
ARG HELM4_SECURITY_PATCH_COMMIT=84e63e5c38913594e476b15609fb6c7ab2d60467
ARG HELM4_APPLIED_PATCH_SHA256=085bde46aa759a7df8a5a7506e5d21f25ccb5ff5b4a47078c21651eaa4673b6e
ARG HELM4_ORAS_GO_VERSION=v2.6.2

ENV DEBIAN_FRONTEND=noninteractive \
    HOME=/home/pasturestack

RUN set -eux; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e "s|http://security.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" \
        -e "s|http://archive.ubuntu.com/ubuntu|${UBUNTU_MIRROR}|g" {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bash \
        bash-completion \
        ca-certificates \
        curl \
        netcat-openbsd \
        util-linux; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble; \
    install -d -m 0750 -o nobody -g nogroup /home/pasturestack; \
    install -d -m 0755 /licenses

COPY --from=kubectl-extractor /out/kubectl /usr/bin/kubectl

RUN set -eux; \
    chmod +x /usr/bin/kubectl; \
    mkdir -p /etc/bash_completion.d; \
    /usr/bin/kubectl completion bash > /etc/bash_completion.d/kubectl

COPY --from=helm2-builder /out/helm /usr/bin/helm
COPY --from=helm4-builder /out/helm4 /usr/bin/helm4
COPY --from=helm4-builder /out/HELM4-LICENSE /licenses/HELM4-LICENSE
COPY --from=helm4-builder /out/HELM4-MODULES.txt /licenses/HELM4-MODULES.txt

RUN chmod 0555 /usr/bin/helm /usr/bin/helm4 && chmod 0444 /licenses/HELM4-LICENSE /licenses/HELM4-MODULES.txt

ENV KUBE_SERVER=http://localhost:8080
ENV SERVER=http://localhost:8080
ENV LISTEN=:8091

COPY kubectl-service kubectl-service.sh kubectl-shell.sh kubectl-shell-entry.sh shell-setup.sh /usr/bin/
COPY update-platform-ca /usr/bin/update-platform-ca
COPY LICENSE SECURITY.md ORIGIN.md COMPATIBILITY.md HELM_PATCH_ORIGIN.md MIGRATION.md targets.lock.env kubernetes-upgrade-path.tsv /licenses/

RUN chmod +x \
        /usr/bin/kubectl-service \
        /usr/bin/kubectl-service.sh \
        /usr/bin/kubectl-shell.sh \
        /usr/bin/kubectl-shell-entry.sh \
        /usr/bin/shell-setup.sh \
        /usr/bin/update-platform-ca

LABEL org.opencontainers.image.title="pasturestack-kubectl-service" \
      org.opencontainers.image.description="PastureStack Kubernetes 1.12 catalog compatibility service" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/kubectl-service" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      io.pasturestack.migration="true" \
      io.pasturestack.helm4.version="${HELM4_TARGET_VERSION}" \
      io.pasturestack.helm4.security-patch="${HELM4_SECURITY_PATCH_COMMIT}" \
      io.pasturestack.helm4.applied-patch-sha256="${HELM4_APPLIED_PATCH_SHA256}" \
      io.pasturestack.helm4.oras-go.version="${HELM4_ORAS_GO_VERSION}"

CMD ["kubectl-service.sh"]

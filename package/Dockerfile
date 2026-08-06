FROM docker:29.6.2-cli@sha256:be132a9f282288de4afaf63379dff75711fda0147c6b72a9df44e51841402144 AS snapshot-ca-certificates

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS helm2-builder

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG GO_VERSION=1.26.5
ARG GO_SHA256_amd64=5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053
ARG HELM_VERSION=v2.17.0
ARG HELM_BUILD_VERSION=v2.17.1
ARG HELM_GIT_COMMIT=a690bad98af45b015bd3da1a41f6218b1a451dbe
ARG HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256=35eafdfed32d4e9612c4f38b65c1d0b3a5e817a81bfea8d3441cf17421deda5b
ARG TILLER_REBUILT_LINUX_AMD64_SHA256=98640ea3be099c5f4cb9dad3c903aa7625092fccd454b88c5e2a0bed48ccfe9a

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
    . /licenses/ubuntu-apt.lock; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        bash="${UBUNTU_APT_BASH_VERSION}" \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        curl="${UBUNTU_APT_CURL_VERSION}" \
        gcc="${UBUNTU_APT_GCC_VERSION}" \
        git="${UBUNTU_APT_GIT_VERSION}" \
        libc6-dev="${UBUNTU_APT_LIBC6_DEV_VERSION}" \
        tar="${UBUNTU_APT_TAR_VERSION}"; \
    mkdir -p /out; \
    { \
        printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; \
        dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; \
    } > /out/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv; \
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
COPY helm-compat-v2.17.0.patch /tmp/helm-lock/helm-compat-v2.17.0.patch

RUN set -eux; \
    mkdir -p /go/src/k8s.io; \
    git clone --depth 1 --branch "${HELM_VERSION}" https://github.com/helm/helm /go/src/k8s.io/helm; \
    cd /go/src/k8s.io/helm; \
    test "$(git rev-parse HEAD)" = "${HELM_GIT_COMMIT}"; \
    echo "${HELM_PATCH_SHA256}  /tmp/helm-lock/helm-compat-v2.17.0.patch" | sha256sum -c -; \
    git apply --check /tmp/helm-lock/helm-compat-v2.17.0.patch; \
    git apply /tmp/helm-lock/helm-compat-v2.17.0.patch; \
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
    env GOCACHE=/tmp/helm2-tiller-build-a GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_BUILD_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=" \
        -o /out/tiller-first ./cmd/tiller; \
    env GOCACHE=/tmp/helm2-tiller-build-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_BUILD_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=" \
        -o /out/tiller-second ./cmd/tiller; \
    cmp /out/tiller-first /out/tiller-second; \
    echo "${TILLER_REBUILT_LINUX_AMD64_SHA256}  /out/tiller-first" | sha256sum -c -; \
    mv /out/tiller-first /out/tiller; \
    rm -f /out/tiller-second; \
    cp LICENSE /out/HELM-LICENSE; \
    ! go version -m /out/tiller | grep -Fq 'github.com/dgrijalva/jwt-go'; \
    go version -m /out/tiller | grep -F 'github.com/prometheus/client_golang' | grep -Fq 'v1.24.1'; \
    /out/tiller --version; \
    cp /tmp/helm-lock/client.go.mod ./go.mod; \
    cp /tmp/helm-lock/client.go.sum ./go.sum; \
    go mod download; \
    go mod verify; \
    env GOCACHE=/tmp/helm2-client-build-a GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_BUILD_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=" \
        -o /out/helm-first ./cmd/helm; \
    env GOCACHE=/tmp/helm2-client-build-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X k8s.io/helm/pkg/version.Version=${HELM_BUILD_VERSION} -X k8s.io/helm/pkg/version.GitCommit=${HELM_GIT_COMMIT} -X k8s.io/helm/pkg/version.BuildMetadata=" \
        -o /out/helm-second ./cmd/helm; \
    cmp /out/helm-first /out/helm-second; \
    echo "${HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256}  /out/helm-first" | sha256sum -c -; \
    mv /out/helm-first /out/helm; \
    rm -f /out/helm-second; \
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
        /tmp/helm2-tiller-build-a \
        /tmp/helm2-tiller-build-b \
        /tmp/helm2-client-build-a \
        /tmp/helm2-client-build-b \
        /tmp/helm-home \
        /root/.cache

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS tiller-runtime

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG TILLER_IMAGE_VERSION=v2.17.1
ARG SOURCE_REVISION=unknown
ARG HELM_VERSION=v2.17.0
ARG HELM_BUILD_VERSION=v2.17.1
ARG TILLER_REBUILT_LINUX_AMD64_SHA256=98640ea3be099c5f4cb9dad3c903aa7625092fccd454b88c5e2a0bed48ccfe9a

ENV DEBIAN_FRONTEND=noninteractive \
    HOME=/home/pasturestack-tiller \
    TILLER_NAMESPACE=kube-system

RUN set -eux; \
    . /licenses/ubuntu-apt.lock; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}"; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble; \
    groupadd --gid 10001 pasturestack-tiller; \
    useradd --uid 10001 --gid 10001 --home-dir /home/pasturestack-tiller --shell /usr/sbin/nologin pasturestack-tiller; \
    install -d -o 10001 -g 10001 /home/pasturestack-tiller /licenses; \
    { \
        printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; \
        dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; \
    } > /licenses/TILLER-UBUNTU-APT-PACKAGES.tsv

COPY --from=helm2-builder /out/tiller /usr/local/bin/tiller
COPY --from=helm2-builder /out/HELM-LICENSE /licenses/HELM-LICENSE
COPY --from=helm2-builder /out/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv /licenses/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv
COPY MIGRATION.md targets.lock.env kubernetes-upgrade-path.tsv /licenses/

RUN set -eux; \
    echo "${TILLER_REBUILT_LINUX_AMD64_SHA256}  /usr/local/bin/tiller" | sha256sum -c -; \
    chmod 0555 /usr/local/bin/tiller; \
    chmod 0444 /licenses/HELM-LICENSE

USER 10001:10001

EXPOSE 44134 44135

LABEL org.opencontainers.image.title="pasturestack-tiller-compat" \
      org.opencontainers.image.description="PastureStack Helm 2.17 Tiller compatibility runtime" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/kubectl-service" \
      org.opencontainers.image.version="${TILLER_IMAGE_VERSION}" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      io.pasturestack.helm2.version="${HELM_BUILD_VERSION}" \
      io.pasturestack.helm2.upstream.version="${HELM_VERSION}" \
      io.pasturestack.ubuntu.snapshot="20260808T000000Z" \
      io.pasturestack.compatibility="helm2"

STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/tiller"]

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS kubectl-extractor

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG KUBERNETES_SOURCE_VERSION=v1.12.10
ARG KUBERNETES_SERVER_SHA512=201ca1af1c4fbecd598a8a6dc0ff5c865e4dab767ec2e768e8100b0c6c450c5ae1ab773aad9e5cdf3f61c15abf3a28f7c4640fbaaac097096905366acbd90a97

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    . /licenses/ubuntu-apt.lock; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        bash="${UBUNTU_APT_BASH_VERSION}" \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        curl="${UBUNTU_APT_CURL_VERSION}" \
        tar="${UBUNTU_APT_TAR_VERSION}"; \
    mkdir -p /out; \
    { \
        printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; \
        dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; \
    } > /out/KUBECTL-EXTRACTOR-UBUNTU-APT-PACKAGES.tsv; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble

RUN set -eux; \
    curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 10 --max-time 300 \
        -o /tmp/kubernetes-server.tgz \
        "https://dl.k8s.io/${KUBERNETES_SOURCE_VERSION}/kubernetes-server-linux-amd64.tar.gz"; \
    echo "${KUBERNETES_SERVER_SHA512}  /tmp/kubernetes-server.tgz" | sha512sum -c -; \
    mkdir -p /out; \
    tar -xzf /tmp/kubernetes-server.tgz -C /out --strip-components=3 \
        kubernetes/server/bin/kubectl; \
    chmod +x /out/kubectl; \
    /out/kubectl version --client=true --short=true; \
    rm -f /tmp/kubernetes-server.tgz

FROM helm2-builder AS helm3-bridge-builder

ARG GO_VERSION=1.26.5
ARG HELM3_BRIDGE_VERSION=v3.21.3
ARG HELM3_BRIDGE_BUILD_VERSION=v3.21.4
ARG HELM3_BRIDGE_GIT_COMMIT=1ad6e68924fdf6fb0c7dcef8e9e1dfc0f36eaed6
ARG HELM3_BRIDGE_SOURCE_SHA256=526f2645c652a3ba2b9b6f38f8d2bab47be4626d4850dafb4311e5f50f6252ef
ARG HELM3_BRIDGE_LICENSE_SHA256=881467e52efeb1807406134b0ec04b1c6f52dc9dc49382713ef9ac4e0cae42f8
ARG HELM3_BRIDGE_APPLIED_PATCH_SHA256=7ea5776330f4af37cc00bb7ea4c2669215fcd5c6b450ca632dc1213c591fe5b2
ARG HELM3_BRIDGE_ORAS_GO_VERSION=v2.6.2
ARG HELM3_BRIDGE_X_TEXT_VERSION=v0.39.0
ARG HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=8c623de98ead4df16c5a95befdd0ddf26a2cd79fe7ea6066f6d44ac1a4d70694

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
        -ldflags "-s -w -X helm.sh/helm/v3/internal/version.version=${HELM3_BRIDGE_BUILD_VERSION} -X helm.sh/helm/v3/internal/version.metadata= -X helm.sh/helm/v3/internal/version.gitCommit=${HELM3_BRIDGE_GIT_COMMIT} -X helm.sh/helm/v3/internal/version.gitTreeState=clean" \
        -o /out/helm3-first ./cmd/helm; \
    rm -rf /tmp/helm3-build-cache-a; \
    env GOCACHE=/tmp/helm3-build-cache-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X helm.sh/helm/v3/internal/version.version=${HELM3_BRIDGE_BUILD_VERSION} -X helm.sh/helm/v3/internal/version.metadata= -X helm.sh/helm/v3/internal/version.gitCommit=${HELM3_BRIDGE_GIT_COMMIT} -X helm.sh/helm/v3/internal/version.gitTreeState=clean" \
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
    test "$(/out/helm3 version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }}')" = "${HELM3_BRIDGE_BUILD_VERSION} ${HELM3_BRIDGE_GIT_COMMIT} clean"; \
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
ARG HELM2TO3_BUILD_VERSION=v0.11.1
ARG HELM2TO3_GIT_COMMIT=98df8bee4094de2e81c0507108867d8139d1306b
ARG HELM2TO3_SOURCE_SHA256=135aab4d90991fe088f3dc1ecf0c1419400d1035068538ef25337c2832080a08
ARG HELM2TO3_LICENSE_SHA256=c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4
ARG HELM2TO3_APPLIED_PATCH_SHA256=d0e5cc51d9ac3e97df4df493e01971f7aa1d9010951aac3edeefc526cc7e3d2d
ARG HELM2TO3_HELM3_VERSION=v3.21.3
ARG HELM2TO3_ORAS_GO_VERSION=v2.6.2
ARG HELM2TO3_X_TEXT_VERSION=v0.39.0
ARG HELM2TO3_REBUILT_LINUX_AMD64_SHA256=111b895b890ade260529fcae29e3b44b7892785391bc39f1b354051ba20807fe

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
        -ldflags "-s -w -X main.version=${HELM2TO3_BUILD_VERSION}" \
        -o /out/helm2to3-first ./main.go; \
    rm -rf /tmp/helm2to3-build-cache-a; \
    env GOCACHE=/tmp/helm2to3-build-cache-b GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
        go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-s -w -X main.version=${HELM2TO3_BUILD_VERSION}" \
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

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS helm-migration-tools-runtime

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG HELM_MIGRATION_IMAGE_VERSION=v0.11.1
ARG SOURCE_REVISION=unknown
ARG HELM3_BRIDGE_VERSION=v3.21.3
ARG HELM3_BRIDGE_BUILD_VERSION=v3.21.4
ARG HELM3_BRIDGE_REBUILT_LINUX_AMD64_SHA256=8c623de98ead4df16c5a95befdd0ddf26a2cd79fe7ea6066f6d44ac1a4d70694
ARG HELM2TO3_VERSION=v0.11.0
ARG HELM2TO3_BUILD_VERSION=v0.11.1
ARG HELM2TO3_REBUILT_LINUX_AMD64_SHA256=111b895b890ade260529fcae29e3b44b7892785391bc39f1b354051ba20807fe

RUN set -eux; \
    . /licenses/ubuntu-apt.lock; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}"; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble; \
    install -d -m 0755 /licenses; \
    { \
        printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; \
        dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; \
    } > /licenses/HELM-MIGRATION-TOOLS-UBUNTU-APT-PACKAGES.tsv

COPY --from=helm3-bridge-builder /out/helm3 /usr/local/bin/helm3
COPY --from=helm3-bridge-builder /out/HELM3-LICENSE /licenses/HELM3-LICENSE
COPY --from=helm3-bridge-builder /out/HELM3-MODULES.txt /licenses/HELM3-MODULES.txt
COPY --from=helm2to3-builder /out/helm2to3 /usr/local/bin/helm2to3
COPY --from=helm2to3-builder /out/HELM2TO3-LICENSE /licenses/HELM2TO3-LICENSE
COPY --from=helm2to3-builder /out/HELM2TO3-MODULES.txt /licenses/HELM2TO3-MODULES.txt
COPY --from=helm2-builder /out/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv /licenses/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv
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
      io.pasturestack.helm3.version="${HELM3_BRIDGE_BUILD_VERSION}" \
      io.pasturestack.helm3.upstream.version="${HELM3_BRIDGE_VERSION}" \
      io.pasturestack.helm2to3.version="${HELM2TO3_BUILD_VERSION}" \
      io.pasturestack.helm2to3.upstream.version="${HELM2TO3_VERSION}" \
      io.pasturestack.execution="explicit-only" \
      io.pasturestack.upstream-binaries="prohibited" \
      io.pasturestack.ubuntu.snapshot="20260808T000000Z"

ENTRYPOINT ["/usr/bin/false"]

FROM helm2-builder AS helm4-builder

ARG GO_VERSION=1.26.5
ARG HELM4_TARGET_VERSION=v4.2.3
ARG HELM4_BUILD_VERSION=v4.2.4
ARG HELM4_TARGET_GIT_COMMIT=43e8b7feece8beb0fcba47059ec9b522fd929a64
ARG HELM4_TARGET_SOURCE_SHA256=9ca555fa3777082fd646c95abf4ee17fc31e25ba97bad76da9de1df2332be168
ARG HELM4_TARGET_LICENSE_SHA256=881467e52efeb1807406134b0ec04b1c6f52dc9dc49382713ef9ac4e0cae42f8
ARG HELM4_SECURITY_PATCH_COMMIT=84e63e5c38913594e476b15609fb6c7ab2d60467
ARG HELM4_SECURITY_PATCH_SHA256=969f510d0f7a6c0b2afad1c1d9b925fd78838923cc1e6b03e01135822230f515
ARG HELM4_APPLIED_PATCH_SHA256=085bde46aa759a7df8a5a7506e5d21f25ccb5ff5b4a47078c21651eaa4673b6e
ARG HELM4_ORAS_GO_VERSION=v2.6.2
ARG HELM4_REBUILT_LINUX_AMD64_SHA256=e472d2b24abf6fcb3a6694e19cd1a8e40cc229193086152cf9011abf3e30a14f

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
        -ldflags "-s -w -X helm.sh/helm/v4/internal/version.version=${HELM4_BUILD_VERSION} -X helm.sh/helm/v4/internal/version.metadata= -X helm.sh/helm/v4/internal/version.gitCommit=${HELM4_TARGET_GIT_COMMIT} -X helm.sh/helm/v4/internal/version.gitTreeState=clean" \
        -o /out/helm4 ./cmd/helm; \
    go version -m /out/helm4 > /out/HELM4-MODULES.txt; \
    cp LICENSE /out/HELM4-LICENSE; \
    echo "${HELM4_TARGET_LICENSE_SHA256}  /out/HELM4-LICENSE" | sha256sum -c -; \
    echo "${HELM4_REBUILT_LINUX_AMD64_SHA256}  /out/helm4" | sha256sum -c -; \
    test "$(awk '$1 == "dep" && $2 == "oras.land/oras-go/v2" {print $3}' /out/HELM4-MODULES.txt)" = "${HELM4_ORAS_GO_VERSION}"; \
    ! grep -Fq 'oras.land/oras-go/v2 v2.6.1' /out/HELM4-MODULES.txt; \
    test "$(/out/helm4 version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }}')" = "${HELM4_BUILD_VERSION} ${HELM4_TARGET_GIT_COMMIT} clean"; \
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

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG IMAGE_VERSION=dev
ARG SOURCE_REVISION=unknown
ARG HELM_VERSION=v2.17.0
ARG HELM_BUILD_VERSION=v2.17.1
ARG HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256=35eafdfed32d4e9612c4f38b65c1d0b3a5e817a81bfea8d3441cf17421deda5b
ARG HELM4_TARGET_VERSION=v4.2.3
ARG HELM4_BUILD_VERSION=v4.2.4
ARG HELM4_SECURITY_PATCH_COMMIT=84e63e5c38913594e476b15609fb6c7ab2d60467
ARG HELM4_APPLIED_PATCH_SHA256=085bde46aa759a7df8a5a7506e5d21f25ccb5ff5b4a47078c21651eaa4673b6e
ARG HELM4_ORAS_GO_VERSION=v2.6.2
ARG HELM4_REBUILT_LINUX_AMD64_SHA256=e472d2b24abf6fcb3a6694e19cd1a8e40cc229193086152cf9011abf3e30a14f

ENV DEBIAN_FRONTEND=noninteractive \
    HOME=/home/pasturestack \
    SSL_CERT_FILE=/home/pasturestack/.local/share/ca-certificates/ca-bundle.crt \
    CURL_CA_BUNDLE=/home/pasturestack/.local/share/ca-certificates/ca-bundle.crt \
    GIT_SSL_CAINFO=/home/pasturestack/.local/share/ca-certificates/ca-bundle.crt \
    KUBECONFIG=/home/pasturestack/.kube/config

RUN set -eux; \
    . /licenses/ubuntu-apt.lock; \
    find /etc/apt -type f \( -name '*.list' -o -name '*.sources' \) -exec sed -i \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' {} +; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::http::Pipeline-Depth "0";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        bash="${UBUNTU_APT_BASH_VERSION}" \
        bash-completion="${UBUNTU_APT_BASH_COMPLETION_VERSION}" \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        curl="${UBUNTU_APT_CURL_VERSION}" \
        netcat-openbsd="${UBUNTU_APT_NETCAT_OPENBSD_VERSION}" \
        util-linux="${UBUNTU_APT_UTIL_LINUX_VERSION}"; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble; \
    install -d -m 0750 -o nobody -g nogroup /home/pasturestack; \
    install -d -m 0755 /licenses; \
    { \
        printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; \
        dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; \
    } > /licenses/KUBECTL-SERVICE-UBUNTU-APT-PACKAGES.tsv

COPY --from=kubectl-extractor /out/kubectl /usr/bin/kubectl

RUN set -eux; \
    chmod +x /usr/bin/kubectl; \
    mkdir -p /etc/bash_completion.d; \
    /usr/bin/kubectl completion bash > /etc/bash_completion.d/kubectl

COPY --from=helm2-builder /out/helm /usr/bin/helm
COPY --from=helm2-builder /out/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv /licenses/HELM2-BUILDER-UBUNTU-APT-PACKAGES.tsv
COPY --from=kubectl-extractor /out/KUBECTL-EXTRACTOR-UBUNTU-APT-PACKAGES.tsv /licenses/KUBECTL-EXTRACTOR-UBUNTU-APT-PACKAGES.tsv
COPY --from=helm4-builder /out/helm4 /usr/bin/helm4
COPY --from=helm4-builder /out/HELM4-LICENSE /licenses/HELM4-LICENSE
COPY --from=helm4-builder /out/HELM4-MODULES.txt /licenses/HELM4-MODULES.txt

RUN set -eux; \
    echo "${HELM_CLIENT_REBUILT_LINUX_AMD64_SHA256}  /usr/bin/helm" | sha256sum -c -; \
    echo "${HELM4_REBUILT_LINUX_AMD64_SHA256}  /usr/bin/helm4" | sha256sum -c -; \
    chmod 0555 /usr/bin/helm /usr/bin/helm4; \
    chmod 0444 /licenses/HELM4-LICENSE /licenses/HELM4-MODULES.txt

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
      io.pasturestack.helm2.version="${HELM_BUILD_VERSION}" \
      io.pasturestack.helm2.upstream.version="${HELM_VERSION}" \
      io.pasturestack.helm4.version="${HELM4_BUILD_VERSION}" \
      io.pasturestack.helm4.upstream.version="${HELM4_TARGET_VERSION}" \
      io.pasturestack.helm4.security-patch="${HELM4_SECURITY_PATCH_COMMIT}" \
      io.pasturestack.helm4.applied-patch-sha256="${HELM4_APPLIED_PATCH_SHA256}" \
      io.pasturestack.helm4.oras-go.version="${HELM4_ORAS_GO_VERSION}" \
      io.pasturestack.ubuntu.snapshot="20260808T000000Z" \
      io.pasturestack.runtime.user="65534:65534" \
      io.pasturestack.shell.requires-privileged="false"

USER 65534:65534
CMD ["kubectl-service.sh"]

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

FROM ubuntu:26.04

ARG UBUNTU_MIRROR=http://archive.ubuntu.com/ubuntu
ARG IMAGE_VERSION=dev
ARG SOURCE_REVISION=unknown

ENV DEBIAN_FRONTEND=noninteractive

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
    rm -f /usr/bin/pebble

COPY --from=kubectl-extractor /out/kubectl /usr/bin/kubectl

RUN set -eux; \
    chmod +x /usr/bin/kubectl; \
    mkdir -p /etc/bash_completion.d; \
    /usr/bin/kubectl completion bash > /etc/bash_completion.d/kubectl

COPY --from=helm2-builder /out/helm /usr/bin/helm

RUN chmod +x /usr/bin/helm

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
      io.pasturestack.migration="true"

CMD ["kubectl-service.sh"]

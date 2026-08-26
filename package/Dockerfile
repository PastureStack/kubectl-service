FROM docker:29.7.2-cli@sha256:000bb62ff495f986c9f5578eb67cc2cb98b91138eda81d7762d5371eb8a497fe AS snapshot-ca-certificates

FROM golang:1.27.0-bookworm@sha256:484ef6066fa69acb059fdfeda7ba2b8f7391f2ef6abc6f9b8411e669ebd56466 AS helm-builder

ARG HELM_VERSION=v4.2.4
ARG HELM_GIT_COMMIT=3900f434fd3ef2b84065dc04508df48f288dba00
ARG HELM_SOURCE_SHA256=f1a1aa56fff071cfa1eecc18c6a7212b2e5c7a0a123fd8e4dcae620bd51b4286
ARG HELM_LICENSE_SHA256=881467e52efeb1807406134b0ec04b1c6f52dc9dc49382713ef9ac4e0cae42f8
ARG HELM_ORAS_GO_VERSION=2.6.2
ARG HELM_PATCHED_GO_MOD_SHA256=3c0166a13121fea5425d0e6ed0f11ad74cc14ed184cf3dfd96724b8113bbbadf
ARG HELM_PATCHED_GO_SUM_SHA256=249a8957dab7b66ea7d71bf9b1e1c8df3cb046a4bff801bdbc2cec210e3f9e21
ENV GOMAXPROCS=2 GOTELEMETRY=off GOTOOLCHAIN=local

RUN set -eux; \
    test "$(go version)" = "go version go1.27.0 linux/amd64"; \
    mkdir -p /src/helm /out; \
    curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location \
        --output /tmp/helm-source.tar.gz "https://codeload.github.com/helm/helm/tar.gz/refs/tags/${HELM_VERSION}"; \
    echo "${HELM_SOURCE_SHA256}  /tmp/helm-source.tar.gz" | sha256sum -c -; \
    tar -xzf /tmp/helm-source.tar.gz -C /src/helm --strip-components=1; \
    rm -f /tmp/helm-source.tar.gz; \
    cd /src/helm; \
    echo "${HELM_LICENSE_SHA256}  LICENSE" | sha256sum -c -; \
    grep -Fxq 'go 1.26.0' go.mod; \
    grep -Fq 'oras.land/oras-go/v2 v2.6.1' go.mod; \
    go mod edit -go=1.27.0 -require="oras.land/oras-go/v2@v${HELM_ORAS_GO_VERSION}"; \
    go mod tidy; \
    echo "${HELM_PATCHED_GO_MOD_SHA256}  go.mod" | sha256sum -c -; \
    echo "${HELM_PATCHED_GO_SUM_SHA256}  go.sum" | sha256sum -c -; \
    go mod verify; \
    test "$(go list -m -f '{{.Version}}' oras.land/oras-go/v2)" = "v${HELM_ORAS_GO_VERSION}"; \
    CGO_ENABLED=0 go build -p=2 -mod=readonly -trimpath -buildvcs=false \
        -ldflags "-w -s -X helm.sh/helm/v4/internal/version.version=${HELM_VERSION} -X helm.sh/helm/v4/internal/version.metadata= -X helm.sh/helm/v4/internal/version.gitCommit=${HELM_GIT_COMMIT} -X helm.sh/helm/v4/internal/version.gitTreeState=clean" \
        -o /out/helm ./cmd/helm; \
    test "$(/out/helm version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }} {{ .GoVersion }}')" \
        = "${HELM_VERSION} ${HELM_GIT_COMMIT} clean go1.27.0"; \
    go version -m /out/helm > /out/HELM-BUILDINFO.txt; \
    awk -v expected="v${HELM_ORAS_GO_VERSION}" \
        '$1 == "dep" && $2 == "oras.land/oras-go/v2" && $3 == expected { found=1 } END { exit found ? 0 : 1 }' \
        /out/HELM-BUILDINFO.txt; \
    cp LICENSE /out/HELM-LICENSE; \
    chmod 0555 /out/helm; \
    chmod 0444 /out/HELM-BUILDINFO.txt /out/HELM-LICENSE

FROM golang:1.27.0-bookworm@sha256:484ef6066fa69acb059fdfeda7ba2b8f7391f2ef6abc6f9b8411e669ebd56466 AS kubectl-builder

ARG KUBERNETES_VERSION=v1.36.4
ARG KUBERNETES_GIT_COMMIT=bb826b1d48562f110659e64e8ec444327433db95
ARG KUBERNETES_SOURCE_SHA256=3c28f11492472df48e658551bf268fd92938b127b0f9dcef7090ac800318c821
ARG KUBERNETES_LICENSE_SHA256=cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30
ENV GOMAXPROCS=2 GOTELEMETRY=off GOTOOLCHAIN=local

RUN set -eux; \
    test "$(go version)" = "go version go1.27.0 linux/amd64"; \
    mkdir -p /src/kubernetes /out; \
    curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location \
        --output /tmp/kubernetes-source.tar.gz "https://dl.k8s.io/${KUBERNETES_VERSION}/kubernetes-src.tar.gz"; \
    echo "${KUBERNETES_SOURCE_SHA256}  /tmp/kubernetes-source.tar.gz" | sha256sum -c -; \
    tar -xzf /tmp/kubernetes-source.tar.gz -C /src/kubernetes; \
    rm -f /tmp/kubernetes-source.tar.gz; \
    cd /src/kubernetes; \
    echo "${KUBERNETES_LICENSE_SHA256}  LICENSE" | sha256sum -c -; \
    GO_VERSION=1.27.0 FORCE_HOST_GO=1 KUBE_BUILD_PLATFORMS=linux/amd64 \
        KUBE_GIT_VERSION="${KUBERNETES_VERSION}" KUBE_GIT_COMMIT="${KUBERNETES_GIT_COMMIT}" \
        KUBE_GIT_TREE_STATE=clean KUBE_GIT_MAJOR=1 KUBE_GIT_MINOR=36 \
        ./hack/make-rules/build.sh cmd/kubectl; \
    install -m 0555 _output/local/go/bin/kubectl /out/kubectl; \
    /out/kubectl version --client=true --output=json | grep -Fq "\"gitVersion\": \"${KUBERNETES_VERSION}\""; \
    go version -m /out/kubectl > /out/KUBECTL-BUILDINFO.txt; \
    grep -Fq '/out/kubectl: go1.27.0' /out/KUBECTL-BUILDINFO.txt; \
    cp LICENSE /out/KUBERNETES-LICENSE; \
    chmod 0444 /out/KUBECTL-BUILDINFO.txt /out/KUBERNETES-LICENSE

FROM ubuntu:26.04@sha256:2260313b31c8c011cd2eebe728008efac1b3982be73eb71348ea2648d2c0e09b

COPY --from=snapshot-ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=helm-builder /out/helm /usr/bin/helm
COPY --from=helm-builder /out/HELM-LICENSE /out/HELM-BUILDINFO.txt /licenses/
COPY --from=kubectl-builder /out/kubectl /usr/bin/kubectl
COPY --from=kubectl-builder /out/KUBERNETES-LICENSE /out/KUBECTL-BUILDINFO.txt /licenses/
COPY ubuntu-apt.lock /licenses/ubuntu-apt.lock

ARG IMAGE_VERSION=dev
ARG SOURCE_REVISION=unknown
ARG KUBERNETES_VERSION=v1.36.4
ARG HELM_VERSION=v4.2.4
ARG HELM_GIT_COMMIT=3900f434fd3ef2b84065dc04508df48f288dba00
ARG HELM_ORAS_GO_VERSION=2.6.2

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
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\nAcquire::AllowInsecureRepositories "false";\nAPT::Get::AllowUnauthenticated "false";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" update; \
    apt-get --snapshot "${UBUNTU_APT_SNAPSHOT}" install -y --no-install-recommends \
        bash="${UBUNTU_APT_BASH_VERSION}" \
        bash-completion="${UBUNTU_APT_BASH_COMPLETION_VERSION}" \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        curl="${UBUNTU_APT_CURL_VERSION}" \
        netcat-openbsd="${UBUNTU_APT_NETCAT_OPENBSD_VERSION}" \
        tar="${UBUNTU_APT_TAR_VERSION}" \
        util-linux="${UBUNTU_APT_UTIL_LINUX_VERSION}"; \
    install -d -m 0755 /licenses /etc/bash_completion.d; \
    install -d -m 0750 -o nobody -g nogroup /home/pasturestack; \
    install -d -m 0700 -o nobody -g nogroup /home/pasturestack/.local/share/ca-certificates; \
    install -m 0600 -o nobody -g nogroup /etc/ssl/certs/ca-certificates.crt "${SSL_CERT_FILE}"; \
    /usr/bin/kubectl completion bash > /etc/bash_completion.d/kubectl; \
    /usr/bin/kubectl version --client=true --output=json | grep -F "${KUBERNETES_VERSION#v}"; \
    test "$(/usr/bin/helm version --template '{{ .Version }} {{ .GitCommit }} {{ .GitTreeState }} {{ .GoVersion }}')" \
        = "${HELM_VERSION} ${HELM_GIT_COMMIT} clean go1.27.0"; \
    { printf 'snapshot\t%s\n' "${UBUNTU_APT_SNAPSHOT}"; dpkg-query -W -f='${binary:Package}\t${Version}\n' | LC_ALL=C sort; } \
        > /licenses/KUBECTL-SERVICE-UBUNTU-APT-PACKAGES.tsv; \
    rm -rf /var/lib/apt/lists/*; \
    rm -f /usr/bin/pebble

ENV KUBE_SERVER=http://localhost:8080 \
    SERVER=http://localhost:8080 \
    LISTEN=:8091

COPY kubectl-service kubectl-service.sh kubectl-shell.sh kubectl-shell-entry.sh shell-setup.sh /usr/bin/
COPY update-platform-ca /usr/bin/update-platform-ca
COPY LICENSE SECURITY.md ORIGIN.md COMPATIBILITY.md /licenses/

RUN chmod 0555 \
        /usr/bin/kubectl-service \
        /usr/bin/kubectl-service.sh \
        /usr/bin/kubectl-shell.sh \
        /usr/bin/kubectl-shell-entry.sh \
        /usr/bin/shell-setup.sh \
        /usr/bin/update-platform-ca

LABEL org.opencontainers.image.title="pasturestack-kubectl-service" \
      org.opencontainers.image.description="PastureStack Kubernetes 1.36 and Helm 4 catalog service" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/kubectl-service" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      io.pasturestack.kubernetes.version="${KUBERNETES_VERSION}" \
      io.pasturestack.helm.version="${HELM_VERSION}" \
      io.pasturestack.tools.compiler="go1.27.0" \
      io.pasturestack.helm.oras-go.version="${HELM_ORAS_GO_VERSION}" \
      io.pasturestack.ubuntu.snapshot="20260808T000000Z" \
      io.pasturestack.runtime.user="65534:65534" \
      io.pasturestack.shell.requires-privileged="false"

USER 65534:65534
CMD ["kubectl-service.sh"]

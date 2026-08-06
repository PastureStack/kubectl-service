#!/bin/bash
set -euo pipefail

umask 077

validate_helm_backend() {
    case "${PASTURESTACK_HELM_BACKEND}" in
        legacy-helm2|helm4)
            ;;
        *)
            echo "unsupported PASTURESTACK_HELM_BACKEND=${PASTURESTACK_HELM_BACKEND}; use legacy-helm2 or helm4" >&2
            return 2
            ;;
    esac
}

run_helm2() {
    /usr/bin/helm "$@"
}

run_helm4() {
    /usr/bin/helm4 "$@"
}

initialize_helm_backend() {
    case "${PASTURESTACK_HELM_BACKEND}" in
        legacy-helm2)
            run_helm2 version --client
            run_helm2 init -c
            ;;
        helm4)
            export HELM_CACHE_HOME="${HOME}/.cache/helm"
            export HELM_CONFIG_HOME="${HOME}/.config/helm"
            export HELM_DATA_HOME="${HOME}/.local/share/helm"
            mkdir -p "${HELM_CACHE_HOME}" "${HELM_CONFIG_HOME}" "${HELM_DATA_HOME}"
            run_helm4 version --short
            ;;
    esac
}

configure_trust() {
    export SSL_CERT_FILE="${SSL_CERT_FILE:-${HOME}/.local/share/ca-certificates/ca-bundle.crt}"
    export CURL_CA_BUNDLE="${CURL_CA_BUNDLE:-${SSL_CERT_FILE}}"
    export GIT_SSL_CAINFO="${GIT_SSL_CAINFO:-${SSL_CERT_FILE}}"
    /usr/bin/update-platform-ca "${SSL_CERT_FILE}"
}

write_kubeconfig() {
    case "${HOME:-}" in
        /*)
            ;;
        *)
            echo 'HOME must be an absolute path' >&2
            return 2
            ;;
    esac

    if [[ ! "${SERVER:-}" =~ ^https?://[^[:space:]]+$ ]]; then
        echo 'SERVER must be an HTTP or HTTPS URL without whitespace' >&2
        return 2
    fi
    local authority=${SERVER#*://}
    authority=${authority%%/*}
    if [[ "${authority}" == *@* ]]; then
        echo 'SERVER must not contain embedded credentials' >&2
        return 2
    fi

    local escaped_server=${SERVER//\\/\\\\}
    escaped_server=${escaped_server//\"/\\\"}
    local kube_directory="${HOME}/.kube"
    local temporary_config
    install -d -m 0700 "${kube_directory}"
    temporary_config=$(mktemp "${kube_directory}/config.tmp.XXXXXX")
    if ! cat > "${temporary_config}" << EOF_CONFIG
apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    server: "${escaped_server}"
  name: "Default"
contexts:
- context:
    cluster: "Default"
  name: "Default"
current-context: "Default"
EOF_CONFIG
    then
        rm -f -- "${temporary_config}"
        return 1
    fi
    chmod 0600 "${temporary_config}"
    mv -f -- "${temporary_config}" "${kube_directory}/config"
    export KUBECONFIG="${KUBECONFIG:-${kube_directory}/config}"
}

main() {
    : "${PASTURESTACK_HELM_BACKEND:=legacy-helm2}"
    validate_helm_backend
    export PASTURESTACK_HELM_BACKEND

    configure_trust
    write_kubeconfig
    initialize_helm_backend

    exec kubectl-service
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi

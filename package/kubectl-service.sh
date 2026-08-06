#!/bin/bash
set -euo pipefail

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
            chown -R nobody:nogroup "${HOME}/.helm"
            ;;
        helm4)
            export HELM_CACHE_HOME="${HOME}/.cache/helm"
            export HELM_CONFIG_HOME="${HOME}/.config/helm"
            export HELM_DATA_HOME="${HOME}/.local/share/helm"
            mkdir -p "${HELM_CACHE_HOME}" "${HELM_CONFIG_HOME}" "${HELM_DATA_HOME}"
            chown -R nobody:nogroup "${HOME}/.cache" "${HOME}/.config" "${HOME}/.local"
            run_helm4 version --short
            ;;
    esac
}

main() {
    : "${PASTURESTACK_HELM_BACKEND:=legacy-helm2}"
    validate_helm_backend
    export PASTURESTACK_HELM_BACKEND

    /usr/bin/update-platform-ca

    mkdir -p "${HOME}/.kube"
    cat > "${HOME}/.kube/config" << EOF_CONFIG
apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    server: "$SERVER"
  name: "Default"
contexts:
- context:
    cluster: "Default"
  name: "Default"
current-context: "Default"
EOF_CONFIG
    chown -R nobody:nogroup "${HOME}/.kube"

    initialize_helm_backend

    exec su -s /bin/bash nobody -p -c "exec kubectl-service"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi

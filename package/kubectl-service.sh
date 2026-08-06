#!/bin/bash
set -euo pipefail

: "${PASTURESTACK_HELM_BACKEND:=legacy-helm2}"
if [[ "${PASTURESTACK_HELM_BACKEND}" != "legacy-helm2" ]]; then
    echo "unsupported PASTURESTACK_HELM_BACKEND=${PASTURESTACK_HELM_BACKEND}; only legacy-helm2 is implemented" >&2
    exit 2
fi
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

helm init -c
chown -R nobody:nogroup "${HOME}/.helm"

exec su -s /bin/bash nobody -p -c "exec kubectl-service"

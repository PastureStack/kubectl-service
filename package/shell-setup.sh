#!/bin/bash
set -euo pipefail

umask 077

token=${KUBECTL_SHELL_TOKEN:-}
if [ -z "${token}" ]; then
    echo 'No shell token provided' >&2
    exit 1
fi
unset KUBECTL_SHELL_TOKEN
set --

session_root=${KUBECTL_SHELL_SESSION_ROOT:-/tmp}
session_root=${session_root%/}
case "${session_root}" in
    /*)
        ;;
    *)
        echo 'KUBECTL_SHELL_SESSION_ROOT must be an absolute path' >&2
        exit 2
        ;;
esac
if [ ! -d "${session_root}" ] || [ ! -w "${session_root}" ]; then
    echo 'The shell session root is missing or not writable' >&2
    exit 1
fi

session_directory=$(mktemp -d "${session_root}/pasturestack-kubectl-shell.XXXXXX")
cleanup() {
    case "${session_directory}" in
        "${session_root}"/pasturestack-kubectl-shell.*)
            rm -rf -- "${session_directory}"
            ;;
        *)
            echo 'Refusing to remove an unexpected shell session path' >&2
            ;;
    esac
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

session_home="${session_directory}/home"
install -d -m 0700 "${session_home}/.kube"
token_base64=$(printf '%s' "${token}" | base64 -w 0)
unset token
kubernetes_url=${KUBERNETES_URL:-https://kubernetes.kubernetes.pasturestack.internal:6443}
if [[ ! "${kubernetes_url}" =~ ^https?://[^[:space:]]+$ ]]; then
    echo 'KUBERNETES_URL must be an HTTP or HTTPS URL without whitespace' >&2
    exit 2
fi
kubernetes_authority=${kubernetes_url#*://}
kubernetes_authority=${kubernetes_authority%%/*}
if [[ "${kubernetes_authority}" == *@* ]]; then
    echo 'KUBERNETES_URL must not contain embedded credentials' >&2
    exit 2
fi
escaped_kubernetes_url=${kubernetes_url//\\/\\\\}
escaped_kubernetes_url=${escaped_kubernetes_url//\"/\\\"}
unset kubernetes_url kubernetes_authority

cat <<EOF > "${session_home}/.kube/config"
apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    certificate-authority: /etc/kubernetes/ssl/ca.pem
    server: "${escaped_kubernetes_url}"
  name: "Default"
contexts:
- context:
    cluster: "Default"
    user: "Default"
  name: "Default"
current-context: "Default"
users:
- name: "Default"
  user:
    token: "${token_base64}"
EOF
unset token_base64 escaped_kubernetes_url

cp /etc/skel/.bashrc "${session_home}/.bashrc"
cat >> "${session_home}/.bashrc" <<EOF
PS1="> "
. /etc/bash_completion
alias k="kubectl"
alias ks="kubectl -n kube-system"
EOF

chmod 0700 "${session_home}/.kube"
chmod 0600 "${session_home}/.kube/config" "${session_home}/.bashrc"

while IFS='=' read -r name _; do
    case "${name}" in
        PLATFORM_ACCESS_KEY|PLATFORM_SECRET_KEY|*_ACCESS_KEY|*_SECRET_KEY|*_CLIENT_SECRET|*_PRIVATE_KEY|*_PASSWORD|*_TOKEN|*_CREDENTIALS)
            unset "${name}"
            ;;
    esac
done < <(env)

cd "${session_home}"
set +e
HOME="${session_home}" \
KUBECONFIG="${session_home}/.kube/config" \
/bin/bash --rcfile "${session_home}/.bashrc" -i
status=$?
set -e
exit "${status}"

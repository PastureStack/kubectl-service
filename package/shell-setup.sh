#!/bin/bash
set -euo pipefail

token=${KUBECTL_SHELL_TOKEN:-${1:-}}

mkdir -p /nonexistent
mount -t tmpfs tmpfs /nonexistent
cd /nonexistent

mkdir -m 700 .kube
cat <<EOF > .kube/config
apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    certificate-authority: /etc/kubernetes/ssl/ca.pem
    server: "${KUBERNETES_URL:-https://kubernetes.kubernetes.pasturestack.internal:6443}"
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
    token: "$(printf '%s' "$token" | base64)"
EOF

cp /etc/skel/.bashrc .
cat >> .bashrc <<EOF
PS1="> "
. /etc/bash_completion
alias k="kubectl"
alias ks="kubectl -n kube-system"
EOF

chown -R nobody:nogroup .kube .bashrc
chmod 700 .kube
chmod 600 .kube/config .bashrc

for i in $(env | cut -d "=" -f 1 | grep "CATTLE\|RANCHER" || true); do
    unset "$i"
done

unset KUBECTL_SHELL_TOKEN token
exec su -s /bin/bash nobody

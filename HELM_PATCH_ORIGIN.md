# Helm Compatibility Patch Origin

The packaging build starts from Helm `v2.17.0` commit
`a690bad98af45b015bd3da1a41f6218b1a451dbe` and applies
`package/helm-v2.17.0-pasturestack.patch`.

The patch:

- replaces two imports from the Kubernetes monolithic module with the matching
  published `k8s.io/kubectl` deployment utilities;
- carries the minimum human-readable table-printer implementation derived from
  Kubernetes `v1.16.2`; and
- updates three dynamic `fmt.Errorf` calls for current Go vet compatibility;
- removes the blanket cloud-provider authentication plug-in import from this
  service-specific client build;
- removes the optional Prometheus interceptors from the maintained Tiller
  compatibility build while preserving the authentication interceptors;
- removes blanket cloud-login plug-in loading from the in-cluster Tiller
  process, which uses its dedicated Kubernetes service account;
- keeps the Tiller `/metrics` endpoint on `prometheus/client_golang` v1.24.1
  with its current compatible dependency set; and
- maps the terminal compatibility surface to `github.com/moby/term` instead of
  linking the complete historical Docker module.

The derived printer files retain their original Kubernetes copyright and
Apache-2.0 license headers. The authoritative sources are:

- <https://github.com/kubernetes/kubernetes/tree/v1.16.2/pkg/printers>
- <https://github.com/kubernetes/kubernetes/tree/v1.16.2/pkg/kubectl/cmd/get>

The full test lock validates the patched Kubernetes-facing package and builds
the Tiller compatibility binary without the retired JWT implementation. The separate client lock contains only
modules required by `cmd/helm`, so server-only modules are not recorded in the
client runtime binary.

This patch removes unnecessary broad dependencies from the client binary. It
does not change the Helm 2 release-storage or Tiller wire protocol. The
service uses the standard kubeconfig token, certificate, and exec credential
paths; provider-specific legacy auth plug-ins are not included.

# PastureStack Kubectl Service

Kubectl Service handles catalog install, upgrade, rollback, remove, list, and
interactive kubectl operations. The maintained runtime uses Kubernetes
`v1.36.4` and Helm `v4.2.4` only.

PastureStack is an independent community project and is not affiliated with
Rancher Labs or SUSE. The repository preserves the history and Apache-2.0
license of the upstream `rancher/kubectld` project.

## Runtime contract

- Helm operations always use Helm `v4.2.4`, rebuilt from its pinned upstream
  source with Go `1.27.0` and patched ORAS `v2.6.2`.
- Kubernetes operations use `kubectl v1.36.4`, rebuilt from the pinned official
  source archive with Go `1.27.0`.
- Catalog event names and reply shapes remain unchanged.
- The former backend selector has been removed. No runtime fallback or
  automatic conversion exists.
- Existing release data must already be readable by Helm 4 before this service
  is introduced. An old release store is not modified or deleted by this image.

The image runs as UID/GID `65534:65534`, needs no Linux capability or privileged
mode, and creates a private trust bundle and per-session shell home.

## Build and test

The service uses Go 1.27 Modules and a checked-in vendor tree. The preserved
Rancher v1 event/client compatibility packages are isolated under
`third_party/`; maintained dependencies such as urfave/cli, Logrus, Gorilla
WebSocket, and `x/sys` are pinned by `go.mod` and `go.sum`.

```sh
make test
make validate
make package IMAGE_NAME=pasturestack/kubectl-service TAG=v1.0.0
```

Packaging verifies both source archives and licenses, the upstream Git commits,
the patched Helm module graph, compiler version, digest-pinned base images, a
non-root read-only smoke run, a client-side Helm render, a Trivy Critical/High
image scan, and a CycloneDX SBOM.

See [COMPATIBILITY.md](COMPATIBILITY.md), [SECURITY.md](SECURITY.md), and
[ORIGIN.md](ORIGIN.md).

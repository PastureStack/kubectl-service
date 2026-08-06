# PastureStack Kubectl Service

Kubectl Service is a compatibility microservice that handles the established catalog create, upgrade, rollback, remove, and query operations by coordinating kubectl and the legacy Helm 2 client contract.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/kubectld`](https://github.com/rancher/kubectld). This GitHub fork preserves the upstream Git history, authorship, dates, tags, and license notices; PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

## Project status

The maintained compatibility candidate is `ghcr.io/pasturestack/kubectl-service:v0.9.11-pasturestack.3`. It uses Ubuntu 26.04, Go 1.26.5, the checksum-verified Kubernetes 1.12.10 kubectl binary, and Helm 2 client and Tiller binaries rebuilt from the pinned source revision with committed module locks. The build container obtains Docker CLI 29.6.2 from the digest-pinned Docker Official Image instead of downloading an unchecked static archive. Both runtime OCI images record the exact source commit in `org.opencontainers.image.revision`; release packaging rejects dirty, untagged, lightweight-tagged, or ambiguously tagged source. The packaging patch removes unnecessary broad dependencies while preserving table output, deployment-wait behavior, release storage, metrics, and the Tiller protocol. The Tiller image runs as UID/GID 10001 and is intended for a read-only root filesystem with all Linux capabilities dropped by the Kubernetes package. Path validation, bounded HTTP, and log-redaction protections are retained. Product build and deployment CI/CD remains disabled while the complete Kubernetes catalog stack is integration-gated; GitHub-managed CodeQL security analysis is enabled with a read-only token and GitHub-owned actions only.

Helm 2 remains a deliberate compatibility boundary and is not a recommendation for new deployments. The compatibility image is versioned `v2.17.0-pasturestack.2`; it is not the final state. The reviewed retirement sequence uses Helm `v3.21.3` only as a conversion bridge and targets Helm `v4.2.3` after the cluster reaches a supported Kubernetes release. See [MIGRATION.md](MIGRATION.md); no migration or Tiller cleanup is automatic.

## Configuration

| Option | Environment | Legacy fallback | Purpose |
| --- | --- | --- | --- |
| `--platform-url` | `PLATFORM_URL` | `CATTLE_URL` | Control-platform API URL. |
| `--platform-access-key` | `PLATFORM_ACCESS_KEY` | `CATTLE_ACCESS_KEY` | API access key. |
| `--platform-secret-key` | `PLATFORM_SECRET_KEY` | `CATTLE_SECRET_KEY` | API secret key. |
| `--worker-count` | `WORKER_COUNT` | none | Event worker count. |
| `--health-check-port` | `HEALTH_CHECK_PORT` | none | Health listener port. |
| `--locale` | `PASTURESTACK_LOCALE` | none | Operator messages: `en-US` or `zh-TW`. |

Legacy names are accepted only as compatibility aliases and are not PastureStack branding. Protocol event names and reply shapes remain unchanged.

Namespace cleanup reads `PLATFORM_KUBERNETES_SERVER`, then the established `SERVER` and `KUBE_SERVER` fallbacks. It accepts only the exact localhost compatibility origin, the established internal Kubernetes service origin, or the standard in-cluster Kubernetes service origins. Arbitrary hosts, embedded credentials, custom proxy paths, and redirects are rejected; namespace values must be valid Kubernetes DNS labels.

## Build and test

The build runs from a Docker-capable Linux host:

```sh
make test
make build
make package IMAGE_NAME=pasturestack/kubectl-service TAG=poc
```

Packaging extracts kubectl from the same checksum-verified Kubernetes 1.12.10 server archive used by the control-plane package. Helm is rebuilt from a shallow checkout of the pinned source revision. `package/helm-v2.17.0.go.mod` and its sum file lock the test graph; `package/helm-v2.17.0-client.go.mod` and its sum file lock the smaller runtime graph. The checksum-pinned patch is applied before either graph is used, and the packaging path does not fetch full dependency Git histories. It does not publish the resulting image. `scripts/check-migration-targets` verifies the reviewed Helm 3 bridge, Helm 4 target, archived migration helper, and all 24 consecutive Kubernetes upgrade checkpoints from 1.13 through 1.36. Offline checks reject skipped minors and malformed locks; online checks compare official stable pointers, Git tag commits, and archive hashes. Downloads are opt-in and never install or execute an artifact. See [COMPATIBILITY.md](COMPATIBILITY.md), [MIGRATION.md](MIGRATION.md), [SECURITY.md](SECURITY.md), [ORIGIN.md](ORIGIN.md), and [HELM_PATCH_ORIGIN.md](HELM_PATCH_ORIGIN.md).

## License and attribution

The inherited project remains licensed under [Apache License 2.0](LICENSE). Copyright and attribution for inherited work and vendored dependencies remain with their respective authors and contributors. PastureStack contributors claim authorship only for their own changes.

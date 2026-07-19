# PastureStack Kubectl Service

Kubectl Service is a compatibility microservice that handles the established catalog create, upgrade, rollback, remove, and query operations by coordinating kubectl and the legacy Helm 2 client contract.

PastureStack is an independent community effort to preserve, audit, and modernize the Rancher 1.6 ecosystem. It is not affiliated with or endorsed by Rancher Labs or SUSE.

**Upstream:** [`rancher/kubectld`](https://github.com/rancher/kubectld). This GitHub fork preserves the upstream Git history, authorship, dates, tags, and license notices; PastureStack maintenance is consolidated into one commit after the preserved upstream boundary.

## Project status

The maintained compatibility candidate uses Ubuntu 26.04, Go 1.26.5, the checksum-verified Kubernetes 1.12.10 kubectl binary, and a Helm 2 client rebuilt from the pinned source revision with committed test and runtime module locks. The packaging patch removes unnecessary broad dependencies from the Helm client while preserving its table output, deployment-wait behavior, and Tiller protocol. Path validation, bounded HTTP, and log-redaction protections are retained. CI/CD remains disabled while the complete Kubernetes catalog stack is integration-gated.

Helm 2 remains a deliberate compatibility boundary and is not a recommendation for new deployments. Replacing it with Helm 3 requires a separate release-storage and Tiller-removal migration.

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

## Build and test

The build runs from a Docker-capable Linux host:

```sh
make test
make build
make package IMAGE_NAME=pasturestack/kubectl-service TAG=poc
```

Packaging extracts kubectl from the same checksum-verified Kubernetes 1.12.10 server archive used by the control-plane package. Helm is rebuilt from a shallow checkout of the pinned source revision. `package/helm-v2.17.0.go.mod` and its sum file lock the test graph; `package/helm-v2.17.0-client.go.mod` and its sum file lock the smaller runtime graph. The checksum-pinned patch is applied before either graph is used, and the packaging path does not fetch full dependency Git histories. It does not publish the resulting image. See [COMPATIBILITY.md](COMPATIBILITY.md), [SECURITY.md](SECURITY.md), [ORIGIN.md](ORIGIN.md), and [HELM_PATCH_ORIGIN.md](HELM_PATCH_ORIGIN.md).

## License and attribution

The inherited project remains licensed under [Apache License 2.0](LICENSE). Copyright and attribution for inherited work and vendored dependencies remain with their respective authors and contributors. PastureStack contributors claim authorship only for their own changes.

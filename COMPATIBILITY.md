# Compatibility

The supported tool contract is Kubernetes `v1.36.4` plus Helm `v4.2.4`.
Rebuilding these tools with Go `1.27.0` does not change their CLI or API
version; the Helm ORAS update affects only OCI artifact transport internals.

Catalog create, upgrade, rollback, delete, and list operations keep their
existing event and response contracts, but releases must be present in the
current Helm storage format. Install and upgrade remain client-side operations
through `--server-side=false`; release listing uses bounded JSON output.

The following obsolete interfaces are intentionally unsupported:

- remote release-server protocols and their cluster Deployment;
- the previous runtime backend-selection flag and environment variable;
- Kubernetes `v1.12` API-server assumptions;
- automatic conversion or deletion of historical release records.

Before rollout, test one representative chart for dependency resolution,
install, upgrade, rollback, delete, list, namespace cleanup, and workload
identity against an isolated Kubernetes `v1.36.4` cluster. That focused matrix
is the required product integration proof; unrelated repositories do not need
to be retested for this change.

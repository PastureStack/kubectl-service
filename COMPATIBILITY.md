# Compatibility Contract

The migration preserves the catalog event names and reply payloads consumed by existing deployments, plus the kubectl create/apply/get behavior and Helm 2 release operations implemented by the upstream service.

The companion Tiller compatibility image preserves Helm 2 ConfigMap or Secret
release records and the v2.17 wire protocol. The Kubernetes package performs a
rolling Deployment update, compares the canonical payload of every release
object before and after the rollout, and restores the prior Deployment revision
if readiness fails or any release is added, removed, or changed. It does not run
`helm init --upgrade`, delete release records, or convert them to Helm 3.

Helm 2 retirement is a separate, operator-controlled migration. The reviewed
targets are recorded in `migration/targets.lock.env`, and the exact consecutive
Kubernetes 1.13-through-1.36 checkpoints are locked in
`migration/kubernetes-upgrade-path.tsv`. These files do not authorize a direct
Kubernetes 1.12-to-current jump, represent EOL checkpoints as supported, or
authorize automatic use of `helm-2to3`. The archived published executable is
not part of the compatibility contract; only the explicit, source-rebuilt,
Critical/High-free migration image may be used. See `MIGRATION.md`
for the backup, per-minor validation, conversion, cleanup, and rollback gates.

The runtime backend is explicit. `PASTURESTACK_HELM_BACKEND=legacy-helm2` is
the default and continues to use the preserved Tiller release store.
`PASTURESTACK_HELM_BACKEND=helm4` is an opt-in post-conversion mode that uses
the separately installed `/usr/bin/helm4` binary, lists JSON releases across
namespaces with a bounded result set, requires an unambiguous namespace for
removal, and preserves client-side apply behavior with
`--server-side=false`. It does not read, convert, delete, or fall back to Helm 2
release records. The binary retains Helm `v4.2.3` behavior but is rebuilt from
the locked source tag with Helm's merged `oras-go` `v2.6.2` security patch. An
invalid backend fails before the health listener or event workers start.

Preferred PastureStack settings use `PLATFORM_URL`, `PLATFORM_ACCESS_KEY`, and `PLATFORM_SECRET_KEY`. The historical `CATTLE_*` settings remain temporary aliases so an existing deployment can be migrated without changing credentials and application code in the same step.

Namespace cleanup prefers `PLATFORM_KUBERNETES_SERVER` and retains `SERVER` and `KUBE_SERVER` as compatibility fallbacks. For SSRF containment, these values select only an exact reviewed internal origin; arbitrary external origins, embedded credentials, custom proxy paths, and redirects are not part of the compatibility contract.

The vendored `github.com/rancher/*` import paths and generated client type names are inherited dependency and wire-schema contracts. They are not product branding and must not be mechanically renamed. Removal requires replacing the generated clients and proving identical event behavior.

Operator messages support `PASTURESTACK_LOCALE=en-US` and `PASTURESTACK_LOCALE=zh-TW`. API payloads, chart content, kubectl output, and Helm output are deliberately not translated.

Before a release, validate Helm 2 and Helm 4 install, upgrade, rollback, remove, list, backend selection, health, malformed or oversized chart paths, secret-redaction, kubectl version, exact Helm versions, and a real catalog event round trip. The separate migration image must additionally prove a non-mutating dry run, preserved Helm 2 release records, an exact workload-object match during conversion, idempotent retry handling, Helm 3 upgrade and rollback, retained object UIDs, retained Tiller, zero exposed test ports, and zero Critical or High image findings.

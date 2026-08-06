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
authorize automatic use of the archived `helm-2to3` plug-in. See `MIGRATION.md`
for the backup, per-minor validation, conversion, cleanup, and rollback gates.

Preferred PastureStack settings use `PLATFORM_URL`, `PLATFORM_ACCESS_KEY`, and `PLATFORM_SECRET_KEY`. The historical `CATTLE_*` settings remain temporary aliases so an existing deployment can be migrated without changing credentials and application code in the same step.

The vendored `github.com/rancher/*` import paths and generated client type names are inherited dependency and wire-schema contracts. They are not product branding and must not be mechanically renamed. Removal requires replacing the generated clients and proving identical event behavior.

Operator messages support `PASTURESTACK_LOCALE=en-US` and `PASTURESTACK_LOCALE=zh-TW`. API payloads, chart content, kubectl output, and Helm output are deliberately not translated.

Before a release, validate install, upgrade, rollback, remove, health, malformed chart paths, secret-redaction, kubectl version, Helm version, and a real catalog event round trip.

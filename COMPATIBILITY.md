# Compatibility Contract

The migration preserves the catalog event names and reply payloads consumed by existing deployments, plus the kubectl create/apply/get behavior and Helm 2 release operations implemented by the upstream service.

Preferred PastureStack settings use `PLATFORM_URL`, `PLATFORM_ACCESS_KEY`, and `PLATFORM_SECRET_KEY`. The historical `CATTLE_*` settings remain temporary aliases so an existing deployment can be migrated without changing credentials and application code in the same step.

The vendored `github.com/rancher/*` import paths and generated client type names are inherited dependency and wire-schema contracts. They are not product branding and must not be mechanically renamed. Removal requires replacing the generated clients and proving identical event behavior.

Operator messages support `PASTURESTACK_LOCALE=en-US` and `PASTURESTACK_LOCALE=zh-TW`. API payloads, chart content, kubectl output, and Helm output are deliberately not translated.

Before a release, validate install, upgrade, rollback, remove, health, malformed chart paths, secret-redaction, kubectl version, Helm version, and a real catalog event round trip.

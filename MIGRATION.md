# Helm 2 Retirement Plan

This repository keeps Helm 2.17 operational only long enough to preserve and
export existing release state. Helm 2 has been unsupported since November 2020,
and the official Helm project warns that vulnerabilities fixed in later releases
will never be repaired in Helm 2. The compatibility runtime is therefore a
temporary isolation boundary, not a supported destination.

## Reviewed targets

The version and checksum source of truth is `migration/targets.lock.env`, reviewed
on 2026-08-07:

- Helm `v3.21.3` is the one-time conversion bridge. Helm 3 security maintenance
  ends on 2027-02-10, so it is not the final state.
- Helm `v4.2.3` is the current stable final target.
- `helm-2to3` `v0.11.0` is the last official conversion helper. Its repository
  is archived and unsupported; use it only from the checksum-verified archive in
  an isolated migration window.
- Kubernetes `v1.36.3` is the current supported target recorded for planning.
  Revalidate it immediately before execution because Kubernetes supports only
  the three newest minor release branches.

Run `scripts/check-migration-targets --online` to compare locks with official
release metadata. Run `scripts/check-migration-targets --download` to download
and hash every Helm migration archive. Neither mode installs or executes an
artifact.

## Required sequence

1. Freeze chart writes and inventory every Helm 2 release, revision, namespace,
   chart, values document, and release storage backend.
2. Run the Kubernetes Package backup tool. Verify its checksums and immediately
   compare the live ConfigMap and Secret canonical payloads with the snapshot.
   Store a second offline copy outside the cluster.
3. Export Kubernetes objects and persistent-data recovery evidence. Identify
   removed Kubernetes APIs in every rendered release before changing the cluster.
4. Upgrade Kubernetes through reviewed, tested intermediate versions. Do not jump
   directly from 1.12 to the current target. At each step, prove API readiness,
   node health, workload continuity, persistent data, networking, rollback, and
   the unchanged Helm 2 snapshot.
5. On a supported Kubernetes version, install the checksum-verified Helm 3 bridge
   outside the control plane. Dry-run each release conversion, convert one release
   without deleting Helm 2 records, and compare rendered manifests, ownership,
   hooks, values, status, and workload identity.
6. Repeat per release. Stop on any unexplained object or payload drift. Retain the
   Helm 2 snapshot and Tiller Deployment until every release passes its rollback
   window.
7. Upgrade the converted Helm 3 state to the checksum-verified Helm 4 target and
   repeat release, hook, values, ownership, workload, data, and rollback checks.
8. Remove Tiller and Helm 2 records only after an explicit operator approval,
   verified offline backup, successful Helm 4 operation, and expiry of the agreed
   rollback window.

## Prohibited shortcuts

- Do not run `helm init --upgrade` against the existing release store.
- Do not use `helm-2to3 --delete-v2-releases` during conversion.
- Do not fetch a plug-in dynamically from the network inside the cluster.
- Do not delete Tiller, ConfigMaps, Secrets, or historical release revisions merely
  because the new client can list a release.
- Do not treat fixture tests as proof of a real release migration.

The current test VM has no authoritative production Helm 2 release set. A real
migration remains incomplete until the sequence above is exercised against an
isolated copy of actual release data and produces preserved-workload and rollback
evidence.

## Authoritative references

- Helm 2 end of support: <https://helm.sh/blog/helm-2-becomes-unsupported/>
- Helm 3 end-of-life schedule: <https://helm.sh/blog/helm-v3-end-of-life/>
- Official Helm releases and signed assets: <https://github.com/helm/helm/releases>
- Archived official Helm 2-to-3 helper: <https://github.com/helm/helm-2to3>
- Supported Kubernetes releases: <https://kubernetes.io/releases/>
- Kubernetes version-skew policy: <https://kubernetes.io/releases/version-skew-policy/>

# Helm 2 Retirement Plan

This repository keeps Helm 2.17 operational only long enough to preserve and
export existing release state. Helm 2 has been unsupported since November 2020,
and the official Helm project warns that vulnerabilities fixed in later releases
will never be repaired in Helm 2. The compatibility runtime is therefore a
temporary isolation boundary, not a supported destination.

## Reviewed targets and Kubernetes path

The target source of truth is `migration/targets.lock.env`. The Kubernetes
minor-by-minor source and artifact lock is
`migration/kubernetes-upgrade-path.tsv`. Both were reviewed on 2026-08-07:

- Helm `v3.21.3` is the one-time conversion bridge. Helm 3 security maintenance
  ends on 2027-02-10, so it is not the final state.
- Helm `v4.2.3` is the current stable final target. The runtime is rebuilt from
  that exact source tag with Helm's subsequently merged two-file
  `oras.land/oras-go/v2` `v2.6.2` patch so the target does not retain High
  `CVE-2026-50163` while awaiting the next upstream release.
- `helm-2to3` `v0.11.0` is the last official conversion helper. Its repository
  is archived and unsupported; use it only from the checksum-verified archive in
  an isolated migration window.
- Kubernetes `v1.36.3` is the current supported target recorded for planning.
  Revalidate it immediately before execution because Kubernetes supports only
  the three newest minor release branches.
- The compatibility source is `v1.12.10-rancher1`, whose upstream-equivalent
  API minor is `v1.12.10`. The path contains exactly 24 consecutive checkpoints,
  from `v1.13.12` through `v1.36.3`; no minor version is omitted.
- Each Kubernetes checkpoint locks the official Git tag commit and the official
  Linux amd64 server archive SHA-256. Versions 1.13 through 1.33 are explicitly
  marked `EOL`; only 1.34, 1.35, and 1.36 are currently marked `SUPPORTED`.
  The EOL checkpoints are unavoidable compatibility waypoints for a no-skip
  control-plane upgrade, not safe long-term operating versions.

Run `scripts/check-migration-targets --offline` to reject skipped or duplicated
Kubernetes minors, malformed commits or hashes, lifecycle drift, and a final
checkpoint that does not match the target. Run it with `--online` to additionally
compare every Kubernetes stable pointer, tag commit, and server archive hash,
confirm that the target is the overall current stable release, derive the oldest
of the three supported minor branches, and compare all Helm locks with official
upstream metadata. `--download` (or
`--download-helm`) downloads and hashes the Helm migration artifacts, including
the exact Helm 4 source archive and reviewed upstream security patch.
`--download-kubernetes VERSION` downloads and verifies one selected locked
Kubernetes server archive; it intentionally does not fetch all 24 large archives
at once. No mode installs, executes, or deploys an artifact.

The lock is a supply-chain and sequencing gate, not a claim that the historical
distribution can already traverse every checkpoint. Before each real step, an
isolated copy of the actual cluster must prove datastore conversion, removed API
handling, networking, admission, storage, node skew, workload identity, rollback,
and distribution-specific upgrade behavior. Stop at the first unsupported or
unexplained result.

## Required sequence

1. Freeze chart writes and inventory every Helm 2 release, revision, namespace,
   chart, values document, and release storage backend.
2. Run the Kubernetes Package backup tool. Verify its checksums and immediately
   compare the live ConfigMap and Secret canonical payloads with the snapshot.
   Store a second offline copy outside the cluster.
3. Export Kubernetes objects and persistent-data recovery evidence. Identify
   removed Kubernetes APIs in every rendered release before changing the cluster.
4. Upgrade Kubernetes through every consecutive checkpoint in
   `migration/kubernetes-upgrade-path.tsv`. Do not jump directly from 1.12 to the
   current target. At each step, use the locked tag and archive evidence, then
   prove API readiness, node health, workload continuity, persistent data,
   networking, rollback, and the unchanged Helm 2 snapshot before proceeding.
5. On a supported Kubernetes version, install the checksum-verified Helm 3 bridge
   outside the control plane. Dry-run each release conversion, convert one release
   without deleting Helm 2 records, and compare rendered manifests, ownership,
   hooks, values, status, and workload identity.
6. Repeat per release. Stop on any unexplained object or payload drift. Retain the
   Helm 2 snapshot and Tiller Deployment until every release passes its rollback
   window.
7. Upgrade the converted Helm 3 state to the checksum-verified Helm 4 target and
   repeat release, hook, values, ownership, workload, data, and rollback checks.
   While catalog writes remain frozen, deploy the reviewed service image with
   `PASTURESTACK_HELM_BACKEND=helm4`, prove list, install, upgrade, rollback, and
   remove behavior against an isolated copy, and confirm that no Helm 2 record
   changed. The runtime does not auto-detect or auto-convert release storage.
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
- Do not treat an official archive checksum as proof that an EOL release is safe
  to operate beyond the shortest controlled upgrade checkpoint.
- Do not set `PASTURESTACK_HELM_BACKEND=helm4` while any catalog-managed release
  still depends on the Helm 2 store. There is no automatic mixed-backend fallback.

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

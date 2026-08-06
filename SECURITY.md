# Security Policy

## Supported state

This repository provides a compatibility candidate for the isolated PastureStack migration environment. Do not deploy a locally modified or unverified image.

## Security boundaries

- Catalog event and reply bodies can contain credentials or chart data and must not be logged.
- Chart paths must remain confined to a newly created mode-0700 temporary directory. Files are created exclusively as mode-0600 regular files; traversal, backslashes, ambiguous root charts, more than 4,096 files, files larger than 8 MiB, and total chart content larger than 64 MiB are rejected before Helm runs.
- Namespace cleanup never sends requests to an arbitrary configured URL. `PLATFORM_KUBERNETES_SERVER`, `SERVER`, and `KUBE_SERVER` are mapped to a fixed set of internal origins; embedded credentials, custom paths, invalid Kubernetes namespace labels, and all HTTP redirects are rejected. Response bodies and untrusted stack names are not copied into operation logs.
- The health listener uses an isolated HTTP mux and bounded server timeouts.
- The runtime CA bootstrap accepts the PastureStack path first; the historical path is a compatibility fallback only.
- The pinned Helm 2.17.0 client and Tiller are a known legacy risk. Both are rebuilt with Go 1.26.5 from a verified source commit and a checksum-pinned patch that removes unnecessary broad dependencies. Tiller does not load the unused cloud-login plug-in bundle, retains `/metrics` through `prometheus/client_golang` v1.24.1, and contains no linked `github.com/dgrijalva/jwt-go` module. Tiller runs as UID/GID 10001 and the Kubernetes package enforces a read-only root filesystem, no privilege escalation, and no Linux capabilities. It must remain limited to the isolated compatibility environment.
- The Kubernetes 1.12.10 kubectl binary is extracted only after the server archive matches the pinned SHA-256 value.
- The build container copies Docker CLI 29.6.2 from the Docker Official Image pinned to multi-platform manifest digest `sha256:be132a9f282288de4afaf63379dff75711fda0147c6b72a9df44e51841402144`. It does not download or execute an unchecked Docker static archive.
- Both release images record the full source commit in the OCI revision label. `scripts/release` requires a clean tree, one annotated PastureStack version tag at `HEAD`, and a tag target equal to the recorded revision. Local dirty candidates remain visibly suffixed with `-dirty` and are not releaseable.
- Helm 3, Helm 4, the archived `helm-2to3` helper, and the current Kubernetes migration target are version- and commit-locked in `migration/targets.lock.env`. All 24 consecutive Kubernetes upgrade checkpoints are additionally version-, commit-, lifecycle-, and archive-hash-locked in `migration/kubernetes-upgrade-path.tsv`. The optional online and download verification modes compare those locks with official release metadata and archive checksums. The archived helper must never be fetched or executed dynamically inside the control plane.
- The Helm 4 runtime is opt-in and never selected by probing cluster state. The official `v4.2.3` source archive, exact tag commit, Apache-2.0 license, upstream security patch commit, upstream patch, and tag-compatible applied patch are independently checksum-locked. Helm had merged the two-file `oras.land/oras-go/v2` `v2.6.2` update after the release; PastureStack applies the same four module-file substitutions and rebuilds with Go 1.26.5 because the published `v4.2.3` binary embeds vulnerable `v2.6.1`. Both patch representations are rejected unless they modify only `go.mod` and `go.sum` and contain the locked old and new ORAS versions. Build-time dependency tests, embedded module evidence, and the release image scan must prove that High `CVE-2026-50163` (`GHSA-fxhp-mv3v-67qp`) is absent. Helm `v4.2.3` is also outside the affected `>=4.0.0, <=4.1.3` ranges for Critical `GHSA-q5jf-9vfq-h4h7` and High `GHSA-vmx8-mqv2-9gmg`. The locked Helm 3 bridge `v3.21.3` is outside the affected ranges for High `GHSA-557j-xg8c-q2mm`. These checks do not authorize plug-in installation or dynamic network fetches inside the control plane.
- Helm 4 release listing is bounded at 10,000 records. Empty or invalid identities and revisions are rejected; deleting a release without a caller-supplied namespace can uninstall the uniquely discovered release but cannot delete that namespace. Invalid backend configuration stops startup before event processing.
- Do not commit API keys, kubeconfigs, certificates, chart secrets, private registry coordinates, or live event payloads.

## Reporting

Report suspected vulnerabilities through this repository's private security advisory channel. Do not include live credentials or production event data in a public issue.

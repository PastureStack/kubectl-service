# Origin and attribution

- Upstream project: `rancher/kubectld`
- Upstream license: Apache-2.0
- Maintained fork: `PastureStack/kubectl-service`

The runtime builds Kubernetes `v1.36.4` kubectl and Helm `v4.2.4` from their
official upstream source archives. Exact source and license checksums plus both
tag commits are pinned in `package/Dockerfile` and `scripts/package`; upstream
licenses and Go build inventories are copied into the release image. The only
dependency override is Helm ORAS `v2.6.1` to fixed `v2.6.2`; its resulting
`go.mod` and `go.sum` hashes are pinned.

PastureStack claims authorship only for its maintenance, integration, security,
and packaging changes. Upstream histories and notices retain their original
attribution.

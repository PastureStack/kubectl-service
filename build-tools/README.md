# Build-tool provenance

The Dapper image treats build tools as reviewed supply-chain inputs. Product
runtime images do not inherit these tools.

## Docker Buildx

- Upstream: <https://github.com/docker/buildx>
- Version: `v0.36.1`
- Source commit: `1d8dde89b8aba914e05e45366770736fea1fd690`
- OCI index digest:
  `sha256:1f2f6b2be4a2511ada67336e76892f1a588c89746009dd4b21069e4d867465be`
- License: Apache-2.0

The build copies `/buildx` only from that digest-pinned upstream image and
checks that the binary reports both the expected version and source commit.
The source-level `go list -mod=vendor -deps ./cmd/buildx` audit records that
[`github.com/docker/docker/pkg/namesgenerator`](buildx-v0.36.1-docker-module-packages.txt)
is the only package compiled from the legacy `github.com/docker/docker`
module. The daemon authorization, archive upload, mount, and `docker cp`
paths associated with the scanner's three High findings are not compiled into
the Buildx CLI. The release-specific assessment is recorded in
[`security/openvex.json`](../security/openvex.json); raw scan output remains
separate, retained evidence.

## Trivy

- Upstream: <https://github.com/aquasecurity/trivy>
- Upstream version: `v0.73.0`
- Rebuilt scanner version: `0.73.1`
- Source commit: `40c73e5d6166dcc0346a1ab4e94499d1572854e4`
- Source archive SHA-256:
  `dd35bec9570e1968b7a3d0d9f6504e5ac2f6a87eea0eee8ddcadd44d08940ee7`
- Source license SHA-256:
  `c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4`
- Compiler: Go `1.26.5`
- Applied patch SHA-256:
  `ae0430b03044dff4b9b2e4115f1e5c156f623356aeb0e1213e4ae0f2c9dd8354`
- Patched ORAS module: `oras.land/oras-go/v2 v2.6.2`
- Patched go-git module: `github.com/go-git/go-git/v5 v5.19.2`
- Patched `go.mod` SHA-256:
  `6b2f30f164c6f377bb4b4cfa06e9b7ece30ddb664bd6a621f5d1518d451c7adc`
- Patched `go.sum` SHA-256:
  `b220e895f26a9234c473be5ad78f624abf3bf49ec30303abbb312029df4d8ce2`
- License: Apache-2.0

The published `v0.73.0` Trivy image is not copied into Dapper. The build
downloads the immutable source archive, verifies its checksum and source
license, applies
[`trivy-v0.73.0-oras-go-v2.6.2-go-git-v5.19.2.patch`](trivy-v0.73.0-oras-go-v2.6.2-go-git-v5.19.2.patch)
with fuzz disabled, verifies the resulting module files, and builds with
`CGO_ENABLED=0`, `GOMAXPROCS=2`, `-p=2`, `-trimpath`, and `-buildvcs=false`.
The explicit parallelism limit prevents the build-only toolchain from
exhausting a small compatibility VM. The Make target passes the same limit to
Dapper for project tests and build orchestration. The embedded Go module
inventory and upstream license are retained under `/licenses` in the Dapper
image.

The patch changes only `go.mod` and `go.sum`. It advances ORAS from `v2.6.1`
to `v2.6.2` and go-git from `v5.19.1` to `v5.19.2`, removing High
`CVE-2026-71556`; it does not change Trivy application source code. Raw scanner
results remain authoritative evidence and are not deleted when a finding is
later classified through a reviewed VEX statement.

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
- Version: `v0.74.0`
- Multi-platform OCI index digest:
  `sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969`
- Vendored upstream license SHA-256:
  `c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4`
- License: Apache-2.0

The Dapper build copies the scanner from the digest-pinned official
multi-platform image, checks the reported version, and retains the upstream
license. Trivy is a build-time scanner and is not copied into the published
Kubectl Service runtime image. No local Trivy fork, source rebuild, or
dependency patch is maintained.

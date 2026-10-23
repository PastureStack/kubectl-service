# Security maintenance

The release image uses Kubernetes `v1.36.4` and Helm `v4.2.4` from immutable,
SHA-256-pinned upstream source archives. Both tools are rebuilt with Go
`1.27.3`; Helm's ORAS dependency is advanced from vulnerable `v2.6.1` to
`v2.6.2`. Helm's version, Git commit, clean tree state, compiler, patched
`go.mod`, and patched `go.sum` are verified during the image build.

Release requirements:

- digest-pinned Ubuntu 26.04 and Docker CLI donor images;
- exact Ubuntu snapshot package versions;
- UID/GID `65534:65534`, read-only-root smoke, no added capability, and
  `no-new-privileges`;
- bounded chart paths and sizes, bounded release listing, and redacted process
  output;
- no available fixed version at any severity, no unreviewed finding, and no
  Critical or High finding in the final image;
- exact official no-fix evidence for residual Low/Medium findings;
- a CycloneDX SBOM for the exact release image.

The removed EOL release-server code, conversion helper, and Kubernetes 1.12
binary are not allowed in production code, package files, or release images.
The validation script enforces this negative boundary.

`security/runtime-openvex.json` applies only to GO-2026-5932 in x/crypto
v0.57.0. Every applicable source build rejects the absent OpenPGP package
before accepting this statement. Raw findings are retained alongside the
effective report; the statement cannot hide another advisory or version.

The non-shipping Docker, Buildx and Trivy verification tools select the
official containerd v2.3.7 and OpenTelemetry v1.45.0 security updates without
changing their own CLI versions. Complete builder reports remain subject to
the same fixed-version policy, with only the existing exact unshipped
kernel-header boundary handled separately.

# Security maintenance

The release image uses Kubernetes `v1.36.4` and Helm `v4.2.4` from immutable,
SHA-256-pinned upstream source archives. Both tools are rebuilt with Go
`1.27.0`; Helm's ORAS dependency is advanced from vulnerable `v2.6.1` to
`v2.6.2`. Helm's version, Git commit, clean tree state, compiler, patched
`go.mod`, and patched `go.sum` are verified during the image build.

Release requirements:

- digest-pinned Ubuntu 26.04 and Docker CLI donor images;
- exact Ubuntu snapshot package versions;
- UID/GID `65534:65534`, read-only-root smoke, no added capability, and
  `no-new-privileges`;
- bounded chart paths and sizes, bounded release listing, and redacted process
  output;
- zero unreviewed Critical or High findings in the final image;
- a CycloneDX SBOM for the exact release image.

The removed EOL release-server code, conversion helper, and Kubernetes 1.12
binary are not allowed in production code, package files, or release images.
The validation script enforces this negative boundary.

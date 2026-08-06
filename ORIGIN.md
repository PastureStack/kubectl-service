# Origin and Attribution

This repository is a GitHub fork of `https://github.com/rancher/kubectld`.

- Preserved upstream boundary: `af773c87d5607f5bdb65a6c36d459c64e865c963`
- Boundary author: Josh Curl
- Boundary date: 2017-09-25
- Migration model: unchanged upstream history followed by one PastureStack maintenance commit

The original commit identifiers, authors, dates, tags, copyright notices, root Apache-2.0 license, and bundled dependency legal files remain authoritative for inherited work. PastureStack claims only its subsequent modifications and does not imply affiliation with the original maintainers.

The optional Helm 4 executable is built from the exact official Helm Project `v4.2.3` source tag. Helm commit `84e63e5c38913594e476b15609fb6c7ab2d60467` updates `oras.land/oras-go/v2` from `v2.6.1` to `v2.6.2`; Helm merged that change after publishing `v4.2.3` but had not published a later release at the 2026-08-07 review. Its upstream patch is checksum-verified. Because that patch's surrounding Kubernetes-module lines come from post-release `main`, packaging applies a separately checksum-locked patch containing the same four ORAS version and module-checksum substitutions against the release tag. Both patch representations are restricted to `go.mod` and `go.sum`. The source archive, module graph, and Apache-2.0 license content are verified independently. The verified license and module inventory are included as `/licenses/HELM4-LICENSE` and `/licenses/HELM4-MODULES.txt`. PastureStack claims authorship only for its packaging and integration changes, not for Helm or the upstream security fix.

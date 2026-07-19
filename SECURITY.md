# Security Policy

## Supported state

This repository provides a compatibility candidate for the isolated PastureStack migration environment. Do not deploy a locally modified or unverified image.

## Security boundaries

- Catalog event and reply bodies can contain credentials or chart data and must not be logged.
- Chart paths must remain confined to the temporary chart directory.
- The health listener uses an isolated HTTP mux and bounded server timeouts.
- The runtime CA bootstrap accepts the PastureStack path first; the historical path is a compatibility fallback only.
- The pinned Helm 2.17.0 client and Tiller are a known legacy risk. The client is rebuilt with Go 1.26.5 from a verified source commit, separate test and runtime module locks, and a checksum-pinned patch that removes unnecessary broad dependencies. It must remain limited to the isolated compatibility environment.
- The Kubernetes 1.12.10 kubectl binary is extracted only after the server archive matches the pinned SHA-256 value.
- Do not commit API keys, kubeconfigs, certificates, chart secrets, private registry coordinates, or live event payloads.

## Reporting

Report suspected vulnerabilities through this repository's private security advisory channel. Do not include live credentials or production event data in a public issue.

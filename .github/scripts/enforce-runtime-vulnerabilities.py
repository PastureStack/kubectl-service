"""Enforce release findings; only exact, reviewed unfixed low/medium entries may pass."""
import argparse
import copy
import json
from pathlib import Path


def build_policy_report(report):
    """Preserve the existing exact, unshipped Linux-header boundary only."""
    filtered = copy.deepcopy(report)
    for result in filtered.get("Results", []):
        result["Vulnerabilities"] = [item for item in result.get("Vulnerabilities") or []
            if not (item.get("PkgName") == "linux-libc-dev"
                    and item.get("InstalledVersion") == "7.0.0-38.38"
                    and not str(item.get("FixedVersion") or "").strip()
                    and item.get("Status") == "affected")]
    return filtered


def violations(report, ledger, repository):
    if (not isinstance(report, dict) or report.get("SchemaVersion") != 2
            or not isinstance(report.get("Results"), list) or not report["Results"]):
        raise ValueError("missing or malformed Trivy scan results")
    for result in report["Results"]:
        if not isinstance(result, dict) or not isinstance(result.get("Target"), str):
            raise ValueError("malformed Trivy result")
        for key in ("Vulnerabilities", "Secrets"):
            if key in result and result[key] is not None and not isinstance(result[key], list):
                raise ValueError("malformed Trivy findings")
    if ledger.get("repository") != repository:
        raise ValueError("allowlist repository mismatch")
    reviewed = {}
    for entry in ledger.get("unfixed", []):
        key = (entry["id"], entry["package"], entry["version"], entry.get("purl", ""))
        if key in reviewed:
            raise ValueError("duplicate reviewed finding")
        if (entry.get("severity") not in ("LOW", "MEDIUM")
                or entry.get("status") != "no_official_fix"
                or not entry.get("officialEvidence", "").startswith("https://")
                or not entry.get("reason") or not entry.get("checkedAt")):
            raise ValueError("incomplete official no-fix evidence")
        reviewed[key] = entry
    errors = []
    for result in report.get("Results") or []:
        if result.get("Secrets"):
            errors.append("secret findings remain")
        for finding in result.get("Vulnerabilities") or []:
            severity = finding.get("Severity", "UNKNOWN")
            key = (finding.get("VulnerabilityID"), finding.get("PkgName"),
                   finding.get("InstalledVersion"),
                   (finding.get("PkgIdentifier") or {}).get("PURL", ""))
            entry = reviewed.get(key)
            fixed = str(finding.get("FixedVersion") or "").strip()
            if (severity not in ("LOW", "MEDIUM") or fixed
                    or entry is None or entry["severity"] != severity):
                errors.append("%s %s@%s severity=%s fixed=%s" % (
                    key[0], key[1], key[2], severity, fixed or "none"))
    return errors


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--phase", choices=("runtime", "builder"), required=True)
    parser.add_argument("--repository", required=True)
    args = parser.parse_args()
    if args.phase == "builder":
        report = build_policy_report(json.loads(Path("dist/dapper.effective.json").read_text(encoding="utf-8")))
        Path("dist/dapper-policy.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    else:
        report = json.loads(Path("dist/kubectl-service.security.json").read_text(encoding="utf-8"))
    failures = violations(report,
                          json.loads(Path(".github/runtime-unfixed.json").read_text(encoding="utf-8")), args.repository)
    if failures:
        raise SystemExit("Release vulnerability policy failed:\n" + "\n".join(failures))
    print("RELEASE_VULNERABILITY_POLICY_OK")

"""Exercise the actual shell guards, including commands after a rejected match."""
import pathlib
import re
import fnmatch
import shlex
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]

# Exact evidence names attested by the v0.9.20 producer, plus its checksum file.
# The SDK inventory glob is retained; every other release asset is explicit.
ATTESTED_ASSETS = {
    'BUILDX-PACKAGES.txt', 'DOCKER-PACKAGES.txt', 'HELM-PACKAGES.txt',
    'KUBECTL-PACKAGES.txt', 'TRIVY-PACKAGES.txt', 'dapper-policy.json',
    'dapper-ubuntu-apt-packages.tsv', 'dapper.accepted-build-header-findings.json',
    'dapper.cdx.json', 'dapper.effective.json', 'dapper.raw.json',
    'kubectl-service-ubuntu-apt-packages.tsv', 'kubectl-service.cdx.json',
    'kubectl-service.security-raw.json', 'kubectl-service.security.json',
    'runtime-openvex.json', 'ubuntu-apt.lock', 'ubuntu-security-packages.lock',
    'ubuntu-backports-evidence.tar', 'ubuntu-backports-validation.json', 'SHA256SUMS',
}


def verify_release_asset_set(workflow):
    lines = workflow.splitlines()
    starts = [i for i, line in enumerate(lines) if line.strip().startswith('gh release create ')]
    assert len(starts) == 1, 'Expected one explicit release publication command'
    command = []
    for line in lines[starts[0]:]:
        command.append(line.strip().removesuffix('\\').strip())
        if not line.rstrip().endswith('\\'):
            break
    patterns = {arg for arg in shlex.split(' '.join(command)) if arg.startswith('dist/')}
    expected = {'dist/' + name for name in ATTESTED_ASSETS}
    assert patterns <= expected | {'dist/*-PACKAGES.txt'}, 'Unexpected/unbounded release asset pattern'
    missing = {name for name in expected if not any(fnmatch.fnmatchcase(name, p) for p in patterns)}
    assert not missing, 'Attested evidence omitted from release: ' + ', '.join(sorted(missing))


class ReleaseAssetSet(unittest.TestCase):
    def test_all_attested_evidence_has_an_explicit_publication_path(self):
        verify_release_asset_set((ROOT / '.github/workflows/release.yml').read_text())

    def test_missing_validation_or_raw_report_is_rejected(self):
        workflow = (ROOT / '.github/workflows/release.yml').read_text()
        for path in ('dist/ubuntu-backports-validation.json', 'dist/kubectl-service.security-raw.json'):
            with self.subTest(path=path), self.assertRaisesRegex(AssertionError, 'Attested evidence omitted'):
                verify_release_asset_set(workflow.replace(path, ''))


class NegativePackageGuards(unittest.TestCase):
    def test_all_package_guards_fail_before_the_next_command(self):
        count = 0
        for relative in ('Dockerfile.dapper', 'package/Dockerfile',
                         '.github/scripts/check-image-policies.sh'):
            lines = (ROOT / relative).read_text().splitlines()
            self.assertFalse(any(re.search(r'^\s*! grep', line) for line in lines))
            for line in lines:
                if 'if grep' not in line:
                    continue
                count += 1
                command = line.strip().removesuffix('\\').strip().rstrip(';')
                command = re.sub(r'(/(?:licenses|out|tmp)/[^; ]+|"dist/\$tool-PACKAGES.txt")',
                                 '"$fixture"', command)
                forbidden = ('dep\tgithub.com/docker/docker\tv1\n'
                             if 'build metadata' in command else
                             'github.com/docker/docker/client\n'
                             if 'upstream Docker dependency' in command else
                             'golang.org/x/crypto/openpgp/packet\n')
                with tempfile.TemporaryDirectory() as temp:
                    fixture = pathlib.Path(temp) / 'packages.txt'
                    for contents, expected in ((forbidden, 1), ('fmt\n', 0)):
                        fixture.write_text(contents)
                        result = subprocess.run(['bash', '-c',
                            'set -euo pipefail; fixture="$1"; ' + command +
                            '; echo NEXT_COMMAND', 'guard-test', str(fixture)],
                            capture_output=True, text=True)
                        self.assertEqual(result.returncode, expected, result.stderr)
                        self.assertEqual('NEXT_COMMAND' in result.stdout, expected == 0)
                        if expected:
                            self.assertIn('Forbidden', result.stderr)
        self.assertEqual(count, 9)


if __name__ == '__main__':
    unittest.main()

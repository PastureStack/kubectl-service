"""Exercise the actual shell guards, including commands after a rejected match."""
import pathlib
import re
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]


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

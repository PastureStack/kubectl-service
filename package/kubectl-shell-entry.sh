#!/bin/bash
set -euo pipefail

exec nc -k -l 10240 > /dev/null 2>&1

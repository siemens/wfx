#!/usr/bin/env bash
set -euo pipefail

apt-get update -q
apt-get install -q -y --no-install-recommends bats jq systemd

#!/usr/bin/env bash
set -euo pipefail

version=$1
checksum=$2
package=/tmp/google-chrome-stable.deb

apt-get update -qq
apt-get upgrade -yqq
curl -fsSL "https://dl.google.com/linux/chrome/deb/pool/main/g/google-chrome-stable/google-chrome-stable_${version}_amd64.deb" -o "$package"
printf '%s  %s\n' "$checksum" "$package" | sha256sum --check
# Keep the installed version pinned even when a later layer upgrades Debian packages.
printf 'repo_add_once="false"\nrepo_reenable_on_distupgrade="false"\n' > /etc/default/google-chrome
DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends "$package"
rm -f /etc/apt/sources.list.d/google-chrome.list
ln -s /usr/bin/google-chrome-stable /usr/bin/chromium

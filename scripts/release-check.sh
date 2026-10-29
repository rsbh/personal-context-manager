#!/usr/bin/env bash
# Checks the snapshot release in dist/ (docs/release.md, Acceptance).
set -euo pipefail

version=$(jq -r .version dist/metadata.json)
targets="darwin_arm64 linux_amd64 linux_arm64"

for t in $targets; do
  archive="dist/pcm_${version}_${t}.tar.gz"
  [ -f "$archive" ] || { echo "missing $archive"; exit 1; }
  files=$(tar -tzf "$archive")
  for f in pcm pcm-server LICENSE README.md; do
    grep -qx "$f" <<<"$files" || { echo "$archive: missing $f"; exit 1; }
  done
  grep -q "pcm_${version}_${t}.tar.gz" dist/checksums.txt || { echo "checksums.txt: missing $archive"; exit 1; }
done

# Run the binary built for this machine, if it is a target.
host="$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/')"
if grep -qw "$host" <<<"$targets"; then
  out=$(mktemp -d)
  tar -xzf "dist/pcm_${version}_${host}.tar.gz" -C "$out" pcm
  line=$("$out/pcm" --version)
  echo "$line"
  case "$line" in
    *"pcm ${version} (commit "*) ;;
    *) echo "unexpected version line: $line"; exit 1 ;;
  esac
  case "$line" in
    *"commit none"*) echo "commit is not set"; exit 1 ;;
  esac
fi
echo "release-check: ok"

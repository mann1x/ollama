#!/usr/bin/env bash
# Publish upstream's MLX runtimes on the fork release for an upstream tag.
#
# xollama serves safetensors models through MLX and must not compile it, so
# the fork supplies upstream's own MLX payloads as pinnable assets, bytes
# unchanged (docs: xollama docs/protocols/FORK-SYNC.md, R6 step 3):
#
#   ollama-linux-amd64-mlx.tar.zst   upstream's asset, republished as is
#   ollama-windows-amd64-mlx.zip     upstream's asset, republished as is
#   ollama-darwin-mlx.tgz            cut from upstream's ollama-darwin.tgz:
#                                    mlx_metal_v3/ (universal), mlx_metal_v4/
#                                    (arm64) and the MLX licence texts
#   mlx-runtime.txt                  provenance and every checksum
#
# usage: scripts/thinkbudget-mlx-assets.sh <upstream tag> <work dir> [--upload]
#   e.g. scripts/thinkbudget-mlx-assets.sh v0.35.1 /path/on/a/real/disk --upload
# The work dir needs about 3 GB and must not be tmpfs. Without --upload the
# assets are built and checked and nothing is published.
set -euo pipefail

up=${1:?upstream tag, e.g. v0.35.1}
work=${2:?work directory}
upload=${3:-}
fork=${FORK_REPO:-mann1x/ollama}
release="${up}-thinkbudget"
here=$(cd "$(dirname "$0")" && pwd)
base="https://github.com/ollama/ollama/releases/download/${up}"

mkdir -p "$work" && cd "$work"
curl -fsSL -o upstream-sha256sum.txt "${base}/sha256sum.txt"
for f in ollama-darwin.tgz ollama-linux-amd64-mlx.tar.zst ollama-windows-amd64-mlx.zip; do
	want=$(awk -v f="./$f" '$2 == f {print $1}' upstream-sha256sum.txt)
	[ -n "$want" ] || { echo "upstream ${up} publishes no $f" >&2; exit 1; }
	if [ ! -s "$f" ] || [ "$(sha256sum "$f" | cut -d' ' -f1)" != "$want" ]; then
		curl -fSL --retry 3 -o "$f" "${base}/$f"
	fi
	got=$(sha256sum "$f" | cut -d' ' -f1)
	[ "$got" = "$want" ] || { echo "$f: sha256 $got, upstream says $want" >&2; exit 1; }
done

python3 "$here/thinkbudget-mlx-darwin.py" ollama-darwin.tgz ollama-darwin-mlx.tgz > darwin-mlx-files.txt

versions=$(cd "$here/.." && for v in MLX_VERSION MLX_C_VERSION; do
	printf '%-14s %s\n' "$v" "$(git show "${up}:${v}" | head -1)"; done)
{
	echo "# MLX runtimes for ${release}: upstream's bytes, republished."
	echo "upstream       ollama/ollama ${up} ($(cd "$here/.." && git rev-parse "${up}^{commit}"))"
	echo "$versions"
	echo
	echo "# assets (sha256, name). The Linux and Windows archives are upstream's"
	echo "# files; their sha256 equals upstream's sha256sum.txt."
	sha256sum ollama-linux-amd64-mlx.tar.zst ollama-windows-amd64-mlx.zip ollama-darwin-mlx.tgz
	echo
	echo "# ollama-darwin-mlx.tgz was cut from upstream's"
	echo "# $(sha256sum ollama-darwin.tgz)"
	echo "# by scripts/thinkbudget-mlx-darwin.py, entry for entry (reproducible)."
	echo "# Its files (sha256, bytes, path), each equal to upstream's:"
	cat darwin-mlx-files.txt
} > mlx-runtime.txt
cat mlx-runtime.txt

if [ "$upload" = "--upload" ]; then
	gh release upload "$release" --repo "$fork" --clobber \
		ollama-linux-amd64-mlx.tar.zst ollama-windows-amd64-mlx.zip \
		ollama-darwin-mlx.tgz mlx-runtime.txt
fi

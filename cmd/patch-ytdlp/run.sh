#!/bin/bash -e
#shellcheck disable=SC2155

export BASE="$(dirname "$(readlink -f "$0")")"

YTDLP_VERSION=${1?:"usage: $0 <version>"}
PATCH_DIR="${BASE}/tmp/${YTDLP_VERSION}"
EXPORT_FILE="${BASE}/export-${YTDLP_VERSION}.json"

# if [ -d "$PATCH_DIR" ]; then
# 	echo "yt-dlp patch already completed for version, not doing anything"
# 	exit 0
# fi

if [ -f "$EXPORT_FILE" ]; then
	echo "yt-dlp export already exists for version at '${EXPORT_FILE}', not doing anything"
	exit 0
fi

echo "patching yt-dlp @ ${YTDLP_VERSION}"

if [ -d "$PATCH_DIR" ]; then
	rm -rf "$PATCH_DIR"
fi

mkdir -vp "$PATCH_DIR"

(
	set -x
	git \
		-c advice.detachedHead=false \
		clone \
		--depth 1 \
		--branch "$YTDLP_VERSION" \
		https://github.com/yt-dlp/yt-dlp.git "$PATCH_DIR"
)

cd "$PATCH_DIR"

if ! grep -q -- "--export-options" "yt_dlp/__main__.py"; then
	(
		set -x
		git apply "${BASE}/export-options.patch"
	)
fi

python3 -m yt_dlp --export-options > "$EXPORT_FILE"

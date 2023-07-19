#!/bin/bash

[ "$DEBUG" == "true" ] && set -x

version=`gh release view -R upfluence/upfluence-if --json tagName -t '{{.tagName}}'`

target_dir="${RUNNER_TOOL_CACHE}/gh-downloader/${version}/${RUNNER_ARCH}"
target_path="${target_dir}/gh-downloader"
force_download=${FORCE_DOWNLOAD:-"false"}

if [ ! -f "$target_path" ] || [ "$force_download" == "true" ]; then
	mkdir -p $target_dir
	curl -L https://github.com/upfluence/gh-downloader/releases/$version/download/gh-downloader-linux-amd64 > $target_path
	chmod +x $target_path
fi

echo $target_dir >> $GITHUB_PATH

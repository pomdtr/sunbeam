#!/bin/sh

# This script installs sunbeam.
#
# Quick install: `curl https://pomdtr.github.io/sunbeam | sh`
#
# Acknowledgments:
#   - getmic.ro: https://github.com/benweissmann/getmic.ro

set -eu

platform=''
machine=$(uname -m)
case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
  "linux")
    case "$machine" in
      "arm64"* | "aarch64"* ) platform='linux_arm64' ;;
      "arm"* | "aarch"*) platform='linux_arm' ;;
      *"86") platform='linux_386' ;;
      *"64") platform='linux_amd64' ;;
    esac
    ;;
  "darwin")
    case "$machine" in
      "arm64"* | "aarch64"* ) platform='darwin_arm64' ;;
      *"64") platform='darwin_amd64' ;;
    esac
    ;;
  *)
    printf "Platform not supported: %s\n" "$(uname -s)"
    exit 1
    ;;
esac

TAG="{{TAG}}"
printf "Target Version: %s\n" "$TAG"
printf "Downloading https://github.com/pomdtr/sunbeam/releases/download/%s/sunbeam-%s.tar.gz\n" "$TAG" "$platform"

temp=$(mktemp -d)
trap 'rm -rf "$temp"' EXIT
curl -L "https://github.com/pomdtr/sunbeam/releases/download/$TAG/sunbeam-$platform.tar.gz" > "$temp/sunbeam.tar.gz"

printf "\nAdding sunbeam binary to %s\n" "/usr/local/bin"

tar zxf "$temp/sunbeam.tar.gz" -C "$temp"
if [ "$(id -u)" -ne 0 ]; then
  sudo mv "$temp/sunbeam" "/usr/local/bin/sunbeam"
else
  mv "$temp/sunbeam" "/usr/local/bin/sunbeam"
fi

printf "\nDone! You can now run sunbeam.\n"

#!/bin/sh

# This script installs Eget.
#
# Quick install: `curl https://pomdtr.github.io/sunbeam | bash`
#
# Acknowledgments:
#   - getmic.ro: https://github.com/benweissmann/getmic.ro

set -e -u

platform=''
machine=$(uname -m)

if [ "${GETSUNBEAM_PLATFORM:-x}" != "x" ]; then
  platform="$GETSUNBEAM_PLATFORM"
else
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
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*)
      printf "Windows is not supported yet\n" >&2
      exit 1
      # case "$machine" in
      #   *"86") platform='windows_386' ;;
      #   *"64") platform='windows_amd64' ;;
      # esac
      ;;
  esac
fi

if [ "x$platform" = "x" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/
Uh oh! We couldn't automatically detect your operating system.
To continue with installation, please choose from one of the following values:
- linux_arm
- linux_arm64
- linux_386
- linux_amd64
- darwin_amd64
- darwin_arm64
Export your selection as the GETSUNBEAM_PLATFORM environment variable, and then
re-run this script.
For example:
  $ export GETSUNBEAM_PLATFORM=linux_amd64
  $ curl https://pomdtr.github.io/sunbeam/install.sh | bash
EOM
  exit 1
else
  printf "Detected platform: %s\n" "$platform"
fi

if [ "x$platform" = "xwindows_amd64" ] || [ "x$platform" = "xwindows_386" ]; then
  extension='zip'
else
  extension='tar.gz'
fi

printf "Latest Version: %s\n" "{{tag}}"
printf "Downloading https://github.com/pomdtr/sunbeam/releases/download/%s/sunbeam-%s.%s\n" "{{tag}}" "$platform" "$extension"

curl -L "https://github.com/pomdtr/sunbeam/releases/download/{{tag}}/sunbeam-$platform.$extension" > "sunbeam.$extension"

case "$extension" in
  "zip") unzip -j "sunbeam.$extension" -d "sunbeam" ;;
  "tar.gz") tar -xvzf "sunbeam.$extension" "sunbeam" ;;
esac

rm "sunbeam.$extension"

cat <<-'EOM'
Eget has been downloaded to the current directory.
You can run it with:
./sunbeam
EOM

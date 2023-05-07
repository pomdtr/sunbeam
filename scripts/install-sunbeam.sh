#!/bin/sh

# This script installs Sunbeam.
#
# Acknowledgments:
#   - getmic.ro: https://github.com/benweissmann/getmic.ro

set -e -u

githubLatestTag() {
  finalUrl=$(curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}')
  printf "%s\n" "${finalUrl##*v}"
}

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
      case "$machine" in
        *"86") platform='windows_386' ;;
        *"64") platform='windows_amd64' ;;
      esac
      ;;
  esac
fi

if [ "x$platform" = "x" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/
Uh oh! We couldn't automatically detect your operating system.
EOM
  exit 1
else
  printf "Detected platform: %s\n" "$platform"
fi

TAG=$(githubLatestTag pomdtr/sunbeam)

if [ "x$platform" = "xwindows_amd64" ] || [ "x$platform" = "xwindows_386" ]; then
  extension='zip'
else
  extension='tar.gz'
fi

printf "Latest Version: %s\n" "$TAG"
printf "Downloading https://github.com/pomdtr/sunbeam/releases/download/v%s/sunbeam-%s-%s.%s\n" "$TAG" "$TAG" "$platform" "$extension"

curl -L "https://github.com/pomdtr/sunbeam/releases/download/v$TAG/sunbeam-$TAG-$platform.$extension" > "sunbeam.$extension"

case "$extension" in
  "zip") unzip -j "sunbeam.$extension" -d "sunbeam-$TAG-$platform" ;;
  "tar.gz") tar -xvzf "sunbeam.$extension" "sunbeam" ;;
esac

rm "sunbeam.$extension"

cat <<-'EOM'
Sunbeam has been downloaded to the current directory.
You can run it with:
./sunbeam
EOM

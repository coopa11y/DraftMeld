#!/usr/bin/env sh
set -eu

repository_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
output_directory="${1:-$repository_root/outputs/release}"
version="$(tr -d '\r\n' < "$repository_root/VERSION")"
temporary_directory="$(mktemp -d)"

cleanup() {
  rm -rf -- "$temporary_directory"
}
trap cleanup EXIT INT TERM

mkdir -p "$output_directory"
rm -f -- \
  "$output_directory/draftmeld_${version}_linux_amd64.tar.gz" \
  "$output_directory/draftmeld_${version}_linux_arm64.tar.gz" \
  "$output_directory/draftmeld_${version}_windows_amd64.zip" \
  "$output_directory/draftmeld_${version}_checksums.txt"

package_target() {
  operating_system="$1"
  architecture="$2"
  executable_name="draftmeld"
  archive_name="draftmeld_${version}_${operating_system}_${architecture}"
  stage_directory="$temporary_directory/$archive_name"

  if [ "$operating_system" = "windows" ]; then
    executable_name="draftmeld.exe"
  fi

  mkdir -p "$stage_directory"
  (
    cd "$repository_root/backend"
    CGO_ENABLED=0 GOOS="$operating_system" GOARCH="$architecture" go build \
      -tags production \
      -trimpath \
      -ldflags="-s -w -X main.version=$version" \
      -o "$stage_directory/$executable_name" \
      ./cmd/draftmeld
  )
  cp "$repository_root/LICENSE" "$repository_root/README.md" "$stage_directory/"

  if [ "$operating_system" = "windows" ]; then
    (cd "$repository_root/backend" && go run ../scripts/zip-directory.go "$stage_directory" "$output_directory/$archive_name.zip")
  else
    tar -C "$temporary_directory" -czf "$output_directory/$archive_name.tar.gz" "$archive_name"
  fi
}

package_target linux amd64
package_target linux arm64
package_target windows amd64

(
  cd "$output_directory"
  sha256sum draftmeld_"$version"_* > "draftmeld_${version}_checksums.txt"
)

echo "Packaged DraftMeld $version release artifacts in $output_directory"

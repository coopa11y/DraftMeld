#!/usr/bin/env sh
set -eu

repository_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(tr -d '\r\n' < "$repository_root/VERSION")"
mkdir -p "$repository_root/outputs"

npm --prefix "$repository_root/frontend" ci
npm --prefix "$repository_root/frontend" run build

cd "$repository_root/backend"
go test ./...
CGO_ENABLED=0 go build -tags production -trimpath -ldflags "-X main.version=$version" -o "$repository_root/outputs/draftmeld" ./cmd/draftmeld

echo "Built DraftMeld $version at $repository_root/outputs/draftmeld"

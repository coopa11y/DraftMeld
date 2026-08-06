$ErrorActionPreference = "Stop"

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$outputDirectory = Join-Path $repositoryRoot "outputs"
$frontendDirectory = Join-Path $repositoryRoot "frontend"
$backendDirectory = Join-Path $repositoryRoot "backend"
$version = (Get-Content -Raw (Join-Path $repositoryRoot "VERSION")).Trim()

New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

npm.cmd --prefix $frontendDirectory ci
npm.cmd --prefix $frontendDirectory run build

Push-Location $backendDirectory
try {
    go test ./...
    go build -tags production -trimpath -ldflags "-X main.version=$version" -o (Join-Path $outputDirectory "draftmeld.exe") ./cmd/draftmeld
}
finally {
    Pop-Location
}

Write-Output "Built DraftMeld $version at $outputDirectory\draftmeld.exe"

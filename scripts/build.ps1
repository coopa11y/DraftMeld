$ErrorActionPreference = "Stop"

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$outputDirectory = Join-Path $repositoryRoot "outputs"
$frontendDirectory = Join-Path $repositoryRoot "frontend"
$backendDirectory = Join-Path $repositoryRoot "backend"

New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null

npm.cmd --prefix $frontendDirectory ci
npm.cmd --prefix $frontendDirectory run build

Push-Location $backendDirectory
try {
    go test ./...
    go build -tags production -trimpath -o (Join-Path $outputDirectory "draftmeld.exe") ./cmd/draftmeld
}
finally {
    Pop-Location
}

Write-Output "Built $outputDirectory\draftmeld.exe"

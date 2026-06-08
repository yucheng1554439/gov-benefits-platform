param(
    [string]$DatabaseUrl = $env:DATABASE_URL
)

$ErrorActionPreference = "Stop"
if (-not $DatabaseUrl) {
    $DatabaseUrl = "postgres://govbenefits:govbenefits@localhost:5432/govbenefits?sslmode=disable"
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

& (Join-Path $ScriptDir "reset_demo_data.ps1") -DatabaseUrl $DatabaseUrl

$SeedFile = Join-Path $ScriptDir "seed_demo_data.sql"
Write-Host "Loading curated demo dataset from $SeedFile ..."

if (Get-Command psql -ErrorAction SilentlyContinue) {
    psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $SeedFile
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    $content = Get-Content $SeedFile -Raw
    $content | docker compose -f (Join-Path $ScriptDir "..\infra\compose\docker-compose.yml") exec -T postgres psql -U govbenefits -d govbenefits -v ON_ERROR_STOP=1
} else {
    Write-Error "Install psql or run Docker with the postgres service available."
}

Write-Host "Demo dataset ready."

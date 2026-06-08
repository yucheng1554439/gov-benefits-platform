# Reset demo transactional data (cases, appeals, audit logs, etc.)
# Preserves agencies, programs, rules, users, and feature flags.

param(
    [string]$DatabaseUrl = $env:DATABASE_URL
)

$ErrorActionPreference = "Stop"

if (-not $DatabaseUrl) {
    $DatabaseUrl = "postgres://govbenefits:govbenefits@localhost:5432/govbenefits?sslmode=disable"
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$SqlFile = Join-Path $ScriptDir "reset_demo_data.sql"

Write-Host "Resetting demo data using $SqlFile ..."

if (Get-Command psql -ErrorAction SilentlyContinue) {
    psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $SqlFile
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    $content = Get-Content $SqlFile -Raw
    $content | docker exec -i gov-benefits-platform-postgres-1 psql -U govbenefits -d govbenefits -v ON_ERROR_STOP=1
    if ($LASTEXITCODE -ne 0) {
        # Try compose service name from infra/compose
        $content | docker compose -f (Join-Path $ScriptDir "..\infra\compose\docker-compose.yml") exec -T postgres psql -U govbenefits -d govbenefits -v ON_ERROR_STOP=1
    }
} else {
    Write-Error "Install psql or run Docker with the postgres service available."
}

Write-Host "Demo data reset complete."

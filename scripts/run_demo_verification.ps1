# Full demo workflow verification against running API.
# Usage: .\scripts\run_demo_verification.ps1

$ErrorActionPreference = "Stop"
$BaseUrl = if ($env:NEXT_PUBLIC_API_URL) { $env:NEXT_PUBLIC_API_URL } else { "http://localhost:8080/api/v1" }
$AgencyId = "22222222-2222-2222-2222-222222222201"
$FoodProgram = "33333333-3333-3333-3333-333333333302"
$Password = "Password123!"

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        $Body = $null,
        [hashtable]$Headers = @{}
    )
    $params = @{
        Uri = "$BaseUrl$Path"
        Method = $Method
        Headers = $Headers
        ContentType = "application/json"
    }
    if ($Body -ne $null) {
        $params.Body = ($Body | ConvertTo-Json -Depth 6)
    }
    return Invoke-RestMethod @params
}

function Login {
    param([string]$Email)
    $tokens = Invoke-Api -Method POST -Path "/auth/login" -Body @{ email = $Email; password = $Password }
    return @{
        Authorization = "Bearer $($tokens.access_token)"
        "X-Agency-ID" = $AgencyId
    }
}

$results = @()

function Test-Step {
    param([string]$Name, [scriptblock]$Action)
    try {
        & $Action
        $script:results += [pscustomobject]@{ Step = $Name; Result = "PASS" }
        Write-Host "[PASS] $Name" -ForegroundColor Green
    } catch {
        $script:results += [pscustomobject]@{ Step = $Name; Result = "FAIL"; Detail = $_.Exception.Message }
        Write-Host "[FAIL] $Name - $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "Demo verification against $BaseUrl"

Test-Step "Citizen login" { $script:citizenHeaders = Login "citizen1@example.com" }
Test-Step "Worker login" { $script:workerHeaders = Login "worker1@dpss.lacounty.gov" }
Test-Step "Supervisor login" { $script:supervisorHeaders = Login "supervisor1@dpss.lacounty.gov" }
Test-Step "Admin login" { $script:adminHeaders = Login "admin@dpss.lacounty.gov" }

Test-Step "Citizen submit application" {
    $case = Invoke-Api -Method POST -Path "/applications" -Headers $script:citizenHeaders -Body @{
        agency_id = $AgencyId
        program_id = $FoodProgram
        household_size = 3
        annual_income = 24000
        employment_status = "employed_part_time"
        zip_code = "90001"
        form_data = @{}
    }
    $script:liveCaseId = $case.id
    if (-not $script:liveCaseId) { throw "No case id returned" }
}

Test-Step "Worker fraud scan" {
    Invoke-Api -Method POST -Path "/cases/$($script:liveCaseId)/fraud/scan" -Headers $script:workerHeaders | Out-Null
}

Test-Step "Worker evaluate eligibility" {
    Invoke-Api -Method POST -Path "/cases/$($script:liveCaseId)/eligibility/evaluate" -Headers $script:workerHeaders | Out-Null
}

Test-Step "Worker calculate benefit" {
    Invoke-Api -Method POST -Path "/cases/$($script:liveCaseId)/benefit/calculate" -Headers $script:workerHeaders | Out-Null
}

Test-Step "Worker move to under review" {
    Invoke-Api -Method PATCH -Path "/cases/$($script:liveCaseId)/status" -Headers $script:workerHeaders -Body @{ to_status = "under_review" } | Out-Null
}

Test-Step "Supervisor deny case" {
    Invoke-Api -Method PATCH -Path "/cases/$($script:liveCaseId)/status" -Headers $script:workerHeaders -Body @{ to_status = "eligibility_review" } | Out-Null
    Invoke-Api -Method PATCH -Path "/cases/$($script:liveCaseId)/status" -Headers $script:supervisorHeaders -Body @{ to_status = "denied" } | Out-Null
}

Test-Step "Citizen file appeal" {
    $appeal = Invoke-Api -Method POST -Path "/appeals" -Headers $script:citizenHeaders -Body @{
        case_id = $script:liveCaseId
        grounds = "Demo appeal: income was miscalculated."
    }
    $script:liveAppealId = $appeal.id
}

Test-Step "Supervisor list pending appeals" {
    $list = Invoke-Api -Method GET -Path "/appeals?pending=true" -Headers $script:supervisorHeaders
    $pending = @($list.data | Where-Object { $_.id -eq $script:liveAppealId })
    if ($pending.Count -eq 0) { throw "Live appeal not in pending queue" }
}

Test-Step "Supervisor decide appeal" {
    Invoke-Api -Method POST -Path "/appeals/$($script:liveAppealId)/decide" -Headers $script:supervisorHeaders -Body @{
        decision = "overturned"
        rationale = "Demo verification approval."
    } | Out-Null
}

Test-Step "Pending queue excludes decided appeal" {
    $list = Invoke-Api -Method GET -Path "/appeals?pending=true" -Headers $script:supervisorHeaders
    $stillPending = @($list.data | Where-Object { $_.id -eq $script:liveAppealId })
    if ($stillPending.Count -gt 0) { throw "Decided appeal still in pending queue" }
}

Test-Step "Duplicate decision blocked" {
    try {
        Invoke-Api -Method POST -Path "/appeals/$($script:liveAppealId)/decide" -Headers $script:supervisorHeaders -Body @{
            decision = "upheld"
            rationale = "Should fail"
        } | Out-Null
        throw "Duplicate decision should have failed"
    } catch {
        if ($_.Exception.Message -notmatch "409|already been decided|Conflict") {
            throw
        }
    }
}

Test-Step "Admin audit trail" {
    $audit = Invoke-Api -Method GET -Path "/audit-logs?limit=10" -Headers $script:adminHeaders
    if (-not $audit.data) { throw "No audit logs returned" }
}

Write-Host "`n--- Summary ---"
$results | Format-Table -AutoSize
$failures = @($results | Where-Object { $_.Result -eq "FAIL" })
if ($failures.Count -gt 0) {
    exit 1
}

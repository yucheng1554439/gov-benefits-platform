# Demo Readiness Report

**Date:** 2026-06-07  
**Target audience:** Hiring managers, LA County SWE interviews, public-sector engineering managers  
**Recording length:** ~5 minutes

---

## Environment

| Item | Status |
|------|--------|
| Docker Compose | Running (`infra/compose/docker-compose.yml`) |
| Backend rebuilt | Yes — includes appeal pending filter + duplicate decision guard |
| Database reset script | `scripts/reset_demo_data.sql` |
| Curated demo seed | `scripts/seed_demo_data.sql` |
| Automated verification | `scripts/run_demo_verification.ps1` — **16/16 PASS** |
| Integration tests | `go test -tags=integration ./tests/integration/...` — **PASS** |

### Quick reset

```powershell
.\scripts\reset_demo_data.ps1          # clean transactional data
.\scripts\seed_demo_data.ps1           # optional curated DEMO-A..D cases
docker compose -f infra/compose/docker-compose.yml up -d --build backend
.\scripts\run_demo_verification.ps1    # verify API workflow
```

---

## Accounts

| Role | Email | Password |
|------|-------|----------|
| Citizen | citizen1@example.com | Password123! |
| Worker | worker1@dpss.lacounty.gov | Password123! |
| Supervisor | supervisor1@dpss.lacounty.gov | Password123! |
| Admin | admin@dpss.lacounty.gov | Password123! |

Agency: LA County DPSS (`22222222-2222-2222-2222-222222222201`)

---

## Verification Results

Executed against **clean database** (reset script) + rebuilt backend.

### Automated API workflow (`run_demo_verification.ps1`)

| Step | Result |
|------|--------|
| Citizen login | **PASS** |
| Worker login | **PASS** |
| Supervisor login | **PASS** |
| Admin login | **PASS** |
| Citizen submit application | **PASS** |
| Worker fraud scan | **PASS** |
| Worker evaluate eligibility | **PASS** |
| Worker calculate benefit | **PASS** |
| Worker move to under review | **PASS** |
| Supervisor deny case | **PASS** |
| Citizen file appeal | **PASS** |
| Supervisor list pending appeals | **PASS** |
| Supervisor decide appeal | **PASS** |
| Pending queue excludes decided appeal | **PASS** |
| Duplicate decision blocked (409) | **PASS** |
| Admin audit trail | **PASS** |

### Integration tests (clean DB)

| Test | Result |
|------|--------|
| Apply → approve → eligibility → benefit → letter download | **PASS** |
| Denied → appeal → supervisor list appeals | **PASS** |

### Feature readiness

| Feature | Demo ready | Evidence |
|---------|------------|----------|
| Application | **Yes** | Verification step + integration test |
| Workflow | **Yes** | Status transitions in verification |
| Eligibility | **Yes** | Evaluate step PASS |
| Benefits | **Yes** | Calculate step PASS |
| Fraud | **Yes** | Scan step PASS |
| Appeals | **Yes** | File, decide, pending filter, duplicate guard PASS |
| Letters | **Yes** | Integration test letter generate + download |
| Audit trail | **Yes** | Admin audit step PASS; events on all actions |

---

## Appeal Workflow Fixes (Phase 2)

| Issue | Fix |
|-------|-----|
| Decided appeals in pending queue | `GET /appeals?pending=true` filters `status=filed`, no decision row, case in `appealed`/`appeal_review` |
| Duplicate decision SQL error | Pre-check + unique violation → `409` "This appeal has already been decided." |
| UI double-submit | Buttons disabled while loading; list refreshes after decision |
| Worker queue misleading | Uses pending API; supervisor decisions on **Supervisor → Appeals** |

---

## Curated Demo Dataset (Phase 4)

| Case | Program | Status | Purpose |
|------|---------|--------|---------|
| CASE-2026-DEMO-A | Food Assistance | Approved | Success path with eval, benefit, letter, audit |
| CASE-2026-DEMO-B | Housing Assistance | Appealed | Live appeal decision on Supervisor → Appeals |
| CASE-2026-DEMO-C | Emergency Relief | Under Review | Open fraud flag |
| CASE-2026-DEMO-D | Healthcare Assistance | Need Documents | Document workflow state |

Load with `scripts/seed_demo_data.ps1` after reset.

---

## UI Cleanup (Phase 5)

| Item | Status |
|------|--------|
| UUIDs in appeals queue | Replaced with case numbers |
| Raw SQLSTATE errors | Appeal duplicate → friendly 409 message |
| Pending appeal filter | Server-side + client-side |
| Supervisor Appeals nav | Added to sidebar |

---

## Remaining Known Issues

### Critical

_None blocking demo recording._

### High

_None — letter generation added to worker case Correspondence card._

### Medium

| Issue | Notes |
|-------|-------|
| Audit log reset requires trigger disable | Handled in `reset_demo_data.sql` (production would archive, not delete) |
| Fresh migration seed includes 2 sample cases | Run reset before demo to remove |

### Low

| Issue | Notes |
|-------|-------|
| Worker Appeals page is read-only | By design — decisions on Supervisor → Appeals |
| Option A letter UI step | May use citizen Letters page after worker generates via case API |

---

## Demo Script

See [demo-script.md](demo-script.md) for click-by-click recording instructions.

---

## Conclusion

The platform is **demo-ready** for a 5-minute recorded walkthrough. Reset scripts provide instant clean state; curated DEMO cases support fast portfolio tours; automated verification confirms the full appeal lifecycle including pending queue correctness and duplicate decision handling.

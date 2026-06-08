# Release Verification

**Release:** Production Readiness RC  
**Date:** 2026-06-07  
**Environment:** Local Docker Compose + integration test suite

---

## Automated Verification

### Integration Tests

```bash
cd backend
go test -tags=integration -v ./tests/integration/...
```

| Test | Result |
|------|--------|
| `TestHappyPath_ApplyToApproval` | **PASS** |
| `TestAppealPath_DeniedToApproved` | **PASS** |

### Unit Tests

```bash
cd backend && go test ./...
```

| Package | Result |
|---------|--------|
| `internal/benefit` | PASS |
| `internal/eligibility` | PASS |
| All other packages | PASS (no test files or cached pass) |

### Health Endpoints

| Endpoint | Expected | Result |
|----------|----------|--------|
| `GET /health` | `{ "status": "ok" }` | **PASS** |
| `GET /ready` | `{ "status": "ready", "checks": { "postgres": "ok", "redis": "ok", "minio": "ok" } }` | **PASS** (when stack running) |

---

## Demo Workflow Verification

Password: `Password123!`

### 1. Citizen → Apply

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Login | `POST /auth/login` | PASS |
| Submit application | `POST /applications` | PASS — case number `CASE-2026-*` |
| View case | `/citizen/cases/{id}` | PASS |

### 2. Worker → Review

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Login worker | `worker1@dpss.lacounty.gov` | PASS |
| Open case | `GET /cases/{id}` | PASS — program name populated |
| Transition status | `PATCH /cases/{id}/status` | PASS |
| Timeline | `GET /cases/{id}/workflow` | PASS — actor names, transitions |

### 3. Worker → Eligibility

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Evaluate | `POST /cases/{id}/eligibility/evaluate` | PASS |
| View result | `GET /cases/{id}/eligibility` | PASS |
| Audit entry | `eligibility.evaluated` in audit log | PASS |

### 4. Worker → Benefit Calculation

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Calculate | `POST /cases/{id}/benefit/calculate` | PASS |
| View amount | `GET /cases/{id}/benefit` | PASS — amount, date, rule version |
| UI | Benefit card shows Recalculate | PASS |

### 5. Supervisor → Approve / Deny

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Login supervisor | `supervisor1@dpss.lacounty.gov` | PASS |
| Approve or deny | `PATCH /cases/{id}/status` | PASS |
| Audit | `case.status_changed` logged | PASS |

### 6. Citizen → Download Letter

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Generate letter | `POST /cases/{id}/letters` (worker) | PASS |
| Download | `GET /letters/{id}/download` | PASS — PDF stream |
| Citizen UI | `/citizen/cases/{id}/letters` | PASS |

### 7. Citizen → File Appeal

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Denied case | Status `denied` | PASS (integration test path) |
| File appeal | `POST /appeals` | PASS |
| Case status | Transitions to `appealed` | PASS |
| Appeals queue | `GET /appeals` — case number, citizen name | PASS |

### 8. Supervisor → Decide Appeal

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Open appeals queue | **`/supervisor/appeals`** (sidebar) | PASS — pending appeals listed |
| Approve appeal | **Approve Appeal** button → `POST /appeals/{id}/decide` (`overturned`) | PASS |
| Deny appeal | **Deny Appeal** button → `POST /appeals/{id}/decide` (`upheld`) | PASS |
| Alt. path | `/supervisor/cases/{id}` → Appeal Review panel | PASS |
| Timeline + audit | Workflow events + `appeal.decided` audit entry | PASS |

**Appeal decision UI routes:**

| Role | Queue (view) | Decision UI |
|------|--------------|-------------|
| Citizen | `/citizen/cases/{id}/appeal` | File appeal only |
| Worker | `/worker/appeals` | Read-only queue |
| Supervisor | **`/supervisor/appeals`** | **Approve Appeal / Deny Appeal** |
| Admin | `/supervisor/appeals` (nav link) | Same as supervisor |

See [appeal-workflow.md](appeal-workflow.md) for full workflow documentation.

### 9. Admin → Audit Review

| Step | Endpoint / UI | Result |
|------|---------------|--------|
| Login admin | `admin@dpss.lacounty.gov` | PASS |
| Audit page | `/admin/audit` | PASS |
| Filters | Search, action filter, pagination | PASS |
| Actor names | Displayed instead of raw IDs | PASS |

---

## Endpoint Verification Summary

| Endpoint | Method | Role | Verified |
|----------|--------|------|----------|
| `/auth/login` | POST | Public | ✅ |
| `/auth/refresh` | POST | Public | ✅ |
| `/applications` | POST | Citizen | ✅ |
| `/cases` | GET | Staff | ✅ |
| `/cases/:id` | GET | All | ✅ |
| `/cases/:id/status` | PATCH | Staff | ✅ |
| `/cases/:id/eligibility/evaluate` | POST | Worker | ✅ |
| `/cases/:id/benefit/calculate` | POST | Worker | ✅ |
| `/cases/:id/letters` | POST | Worker | ✅ |
| `/letters/:id/download` | GET | All | ✅ |
| `/appeals` | POST | Citizen | ✅ |
| `/appeals` | GET | Staff | ✅ |
| `/appeals/:id/decide` | POST | Supervisor | ✅ |
| `/audit-logs` | GET | Admin/Supervisor | ✅ |
| `/admin/eligibility-rules` | GET | Admin | ✅ |
| `/admin/eligibility-rules/:id/simulate` | POST | Admin | ✅ |
| `/health` | GET | Public | ✅ |
| `/ready` | GET | Public | ✅ |

---

## Screenshots

Manual captures recommended for portfolio (store under `docs/screenshots/`):

1. `citizen-application.png` — Application wizard
2. `worker-review.png` — Case detail with actions
3. `eligibility-evaluation.png` — Eligibility card with trace
4. `benefit-calculation.png` — Benefit amount with rule version
5. `appeals-queue.png` — Worker appeals with case numbers
6. `audit-trail.png` — Admin audit with filters
7. `analytics-dashboard.png` — Supervisor analytics

---

## Result

**All demo workflow steps: PASS**  
**Integration test suite: PASS**  
**Release candidate: APPROVED for portfolio deployment**

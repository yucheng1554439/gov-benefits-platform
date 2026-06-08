# Release Candidate Plan (RC1)

**Source of truth:** [`docs/implementation-audit.md`](implementation-audit.md)  
**Target:** Portfolio-quality demo — **not** 100% feature completion  
**Demo script:** Citizen Application → Eligibility → Benefit Calculation → Worker Review → Supervisor Approval → Letter Generation → Appeal → Audit Trail

## Scope Rules

### In scope (demo workflow)
Fix bugs and gaps that block the scripted demo end-to-end.

### Out of scope (ignored NOT IMPLEMENTED items)
These are explicitly **deferred** for RC1:

| Feature | Reason |
|---------|--------|
| Eligibility simulation | Not on demo path |
| CSV / Excel / PDF reporting exports | Not on demo path |
| Email notifications | Audit trail uses `audit_logs`, not email |
| Backend feature-flag enforcement | UI gating sufficient for demo |
| Background jobs (SLA cron, retention, notification queue, benefit recalc) | Sync APIs cover demo |
| Agency switching | Single-agency demo accounts |
| Document upload UI | Not on demo script |
| Admin rule/template CRUD | Seed data sufficient |

---

## Demo Workflow Gap Analysis

| Step | Pre-RC Status | Blocker |
|------|---------------|---------|
| Citizen Application | COMPLETE | None critical |
| Eligibility | API only | No worker UI to evaluate/display |
| Benefit Calculation | BROKEN UI | `monthly_amount` vs `calculated_amount` |
| Worker Review | PARTIAL | Missing program name on case GET |
| Supervisor Approval | COMPLETE | Works via status transitions |
| Letter Generation | BROKEN | MinIO `NoSuchBucket`; no download |
| Appeal | BROKEN/PARTIAL | Case workflow not updated; no agency appeals list |
| Audit Trail | PARTIAL | API exists; no viewer UI |

---

## Top 10 Fixes (Ranked by Demo Impact)

| Rank | Fix | Demo step unlocked | Effort | Owner layer |
|------|-----|-------------------|--------|-------------|
| **1** | **MinIO bucket bootstrap on startup** | Letter Generation | **S (1h)** | Backend infra |
| **2** | **Benefit UI field mapping** (`calculated_amount`) | Benefit Calculation | **XS (30m)** | Frontend |
| **3** | **Eligibility evaluate + display on worker case page** | Eligibility | **M (2h)** | Frontend |
| **4** | **Letter download endpoint + citizen letters UI** | Letter Generation | **M (3h)** | Backend + Frontend |
| **5** | **Appeal ↔ workflow integration** (file → `appealed`; decide → case status) | Appeal | **M (4h)** | Backend |
| **6** | **Join program on `CaseService.Get`** | Worker Review | **S (1h)** | Backend |
| **7** | **Audit trail viewer page** (supervisor/admin) | Audit Trail | **M (2h)** | Frontend |
| **8** | **`GET /appeals` agency endpoint** | Appeal queue / supervisor review | **S (1h)** | Backend |
| **9** | **Appeal review panel status filter** | Appeal | **XS (30m)** | Frontend |
| **10** | **Repair integration tests** (demo workflow contract) | RC verification gate | **M (2h)** | Tests |

**Total estimated effort:** ~17 hours

---

## RC1 Acceptance Criteria

After fixes 1–10:

1. Citizen can submit application and land on case detail.
2. Worker opens case → runs eligibility evaluate → sees eligible/ineligible + trace.
3. Worker calculates benefit → **amount displays** on case page.
4. Worker transitions case through review; program name visible.
5. Supervisor approves or denies case.
6. Worker/supervisor generates approval or denial letter → **PDF stored and downloadable** by citizen.
7. Citizen files appeal on denied case → case status becomes `appealed`.
8. Supervisor reviews appeal → decision updates case workflow status.
9. Supervisor/admin opens audit viewer → sees `case.status_changed` and related events.
10. `go test -tags=integration ./tests/integration/...` passes against running API.

---

## Verification Plan

1. Rebuild Docker Compose (`backend`, `frontend`, `worker`).
2. Run scripted demo manually (accounts from seed).
3. Run integration tests.
4. Regenerate audit → [`docs/implementation-audit-v2.md`](implementation-audit-v2.md).

---

## Post-RC Backlog (not in RC1)

- Token refresh on frontend
- Real agency switching (`X-Agency-ID` or reissue JWT)
- Notification `user_id` fix
- Report file generation
- Document upload UI
- SLA at-risk detection
- Feature flag backend enforcement

---

## Implementation Status

| # | Fix | Status |
|---|-----|--------|
| 1 | MinIO bucket bootstrap | Implemented |
| 2 | Benefit UI mapping | Implemented |
| 3 | Eligibility worker UI | Implemented |
| 4 | Letter download | Implemented |
| 5 | Appeal workflow | Implemented |
| 6 | Case program join | Implemented |
| 7 | Audit viewer | Implemented |
| 8 | GET /appeals | Implemented |
| 9 | Appeal panel filter | Implemented |
| 10 | Integration tests | Implemented |

See [`docs/implementation-audit-v2.md`](implementation-audit-v2.md) for post-RC verification results.

# Program Display Fix

**Issue:** Worker Queue (and citizen dashboard) showed `Program = "—"` for all cases in production.

**Example (before fix):**

| Case # | Program |
|--------|---------|
| CASE-2026-000001 | — |

**Expected:**

| Case # | Program |
|--------|---------|
| CASE-2026-000001 | Food Assistance |

---

## Root cause

The UI renders `row.program?.name ?? '—'` in `frontend/src/app/worker/queue/page.tsx`. The data comes from `GET /api/v1/cases` via `useCases()`.

| Endpoint | Program populated? |
|----------|-------------------|
| `GET /cases/:id` | **Yes** — `CaseService.Get` loads program via `agencies.GetProgramByID` |
| `GET /cases` (list) | **No** — `CaseService.List` returned raw case rows without program enrichment |

Production verification (2026-06-08):

```json
// GET /cases — first row
{ "case_number": "CASE-2026-000001", "program_id": "33333333-...", "program": null }

// GET /cases/{id}
{ "program": { "name": "Food Assistance", "code": "food_assistance" } }
```

The list handler never joined or hydrated program names, so every queue row fell back to the em dash placeholder.

---

## Fix

**File:** `backend/internal/service/case.go`

After loading cases in `List`, enrich each row with the same program lookup used by `Get`:

```go
for i := range cases {
    if cases[i].ProgramID == uuid.Nil {
        continue
    }
    program, err := s.agencies.GetProgramByID(ctx, cases[i].ProgramID)
    if err != nil || program == nil {
        continue
    }
    cases[i].Program = program
}
```

No frontend change required — the queue already displays `program.name` when present.

---

## Verification

### Local (after rebuild/redeploy backend)

```powershell
# Login as worker, then:
curl -H "Authorization: Bearer $TOKEN" `
     -H "X-Agency-ID: 22222222-2222-2222-2222-222222222201" `
     https://<api-host>/api/v1/cases
```

Each item should include:

```json
"program": { "name": "Food Assistance", ... }
```

### Production

Redeploy Render backend with this commit, then confirm Worker Queue at  
https://gov-benefits-platform.vercel.app/worker/queue shows program names.

---

## Related surfaces fixed by the same API change

- Worker Queue (`/worker/queue`)
- Citizen Dashboard (`/citizen/dashboard`)
- Supervisor Escalations case table (uses same `useCases` hook)

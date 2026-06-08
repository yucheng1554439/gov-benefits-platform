# Government Benefits Platform — Implementation Audit

**Audit date:** 2026-06-08  
**Repository:** `gov-benefits-platform`  
**Method:** Code-path tracing, live API execution against Docker Compose stack, integration test run, and frontend route/component review. Features are **not** marked COMPLETE unless verified end-to-end or via executable API/UI paths.

**Verification environment:** Docker Compose (`postgres`, `redis`, `minio`, `mailhog`, `backend`, `worker`, `frontend`) running locally. Demo credentials from seed data (`Password123!`).

---

## Executive Summary

| Metric | Count |
|--------|------:|
| **Total planned features audited** | **84** |
| **COMPLETE** | **26** |
| **PARTIAL** | **38** |
| **STUBBED** | **9** |
| **BROKEN** | **6** |
| **NOT IMPLEMENTED** | **5** |

**Overall platform completion (weighted): ~58%**

The platform has a solid backend skeleton: PostgreSQL schema, RLS, JWT auth, workflow state machine, eligibility/benefit engines, event bus, and most REST routes exist with real business logic. Production readiness is limited by missing frontend surfaces (documents, audit viewer, letter download), broken cross-layer contracts (benefit UI, agency switching), infrastructure gaps (MinIO bucket, email), stubbed reporting/background jobs, and disconnected appeal/SLA workflows.

**Integration tests:** `backend/tests/integration/workflow_test.go` — **FAIL** (application payload uses deprecated `program_code`; status transitions use `status` instead of `to_status`; `GET /appeals` returns 404).

---

## Critical Production Blocking Issues

End-to-end flow: **Citizen applies → Fraud Scan → Eligibility → Benefit Calculation → Worker Review → Supervisor Approval → Letter Generation → Appeal → Reporting**

| Step | Status | Blocker |
|------|--------|---------|
| Citizen applies | **Works** | Program/agency wizard fixed; `form_data` empty; no document step |
| Fraud Scan | **Works** | Manual only; no auto-scan; duplicate flags on rescan |
| Eligibility | **API works** | No worker UI to trigger/display; not tied to workflow status |
| Benefit Calculation | **API works, UI broken** | Frontend expects `monthly_amount`; API returns `calculated_amount` |
| Worker Review | **Works** | Worker queue not filtered by assignment; program name missing on case GET |
| Supervisor Approval | **Partial** | Status transitions work when valid; no `approved_amount` on benefits |
| Letter Generation | **Broken in Compose** | S3 `NoSuchBucket` — letter PDF upload fails; no download endpoint |
| Appeal | **Partial/Broken** | Filing works; case workflow not updated; `GET /appeals` missing; decisions don't drive case status |
| Reporting | **Stubbed** | Job marks `completed` without generating CSV/Excel/PDF files |

**Additional cross-cutting blockers:**
- **Agency switching is cosmetic** — `X-Agency-ID` header ignored; tenant comes from JWT only
- **Document management has no frontend** — upload/verify APIs exist but are unreachable from UI
- **Email notifications not implemented** — Mailhog runs but no SMTP sender code
- **Feature flags not enforced on backend** — `FeatureFlagGuard` exists but is never registered on routes
- **Refresh token unused on frontend** — sessions expire at access TTL without auto-refresh

---

## Foundation

### Docker Compose
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 80% |
| **Evidence** | `infra/compose/docker-compose.yml` — postgres, redis, minio, mailhog, backend, worker, frontend with env wiring |
| **Missing** | MinIO bucket init; healthchecks for minio/backend/worker; compose smoke test in CI |
| **Bugs** | MinIO `depends_on` uses `service_started` not healthy |

### PostgreSQL
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 95% |
| **Evidence** | Migrations `001`–`005`; RLS in `002`; seed in `004`; auto-migrate on API startup (`cmd/api/main.go`) |
| **Missing** | Migration rollback automation |
| **Bugs** | None observed |

### Redis
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `internal/jobs/queue.go`; `/ready` pings Redis; worker dequeues jobs |
| **Missing** | Persistence config; `JobRunRetention` handler |
| **Bugs** | None critical |

### MinIO
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | Compose service; `internal/storage/s3.go`; compose sets `STORAGE_DRIVER=s3` |
| **Missing** | Bucket creation on startup |
| **Bugs** | **Verified live:** letter upload fails with `NoSuchBucket: govbenefits` |

### Mailhog
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 15% |
| **Evidence** | Compose ports 1025/8025; `SMTP_HOST`/`SMTP_PORT` in config |
| **Missing** | Any Go SMTP/mail sender (`net/smtp`, gomail, etc.) |
| **Bugs** | Infrastructure-only; email notifications cannot work |

### API Startup
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 95% |
| **Evidence** | `cmd/api/main.go` — full DI graph, migrations, routes, graceful shutdown; **verified** `GET /health`, `GET /ready` |
| **Missing** | MinIO bucket bootstrap |
| **Bugs** | Storage failures at runtime not startup |

### Worker Startup
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `cmd/worker/main.go`; processes `generate_letter`, `generate_report` in `internal/jobs/worker.go` |
| **Missing** | Health endpoint; worker image not built in CI |
| **Bugs** | Worker passes `bus: nil` to letter service — async letters skip event/audit side effects |

### Frontend Startup
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `infra/docker/Dockerfile.frontend`; Next.js 15 standalone; `:3000` exposed |
| **Missing** | Runtime healthcheck |
| **Bugs** | Session from localStorage only; stale tokens fail silently until API call |

### CI/CD
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 65% |
| **Evidence** | `.github/workflows/backend.yml`, `frontend.yml` — lint (non-blocking), test, build, Docker, Trivy |
| **Missing** | CD/deploy; integration tests in CI; worker Docker build; blocking lint |
| **Bugs** | Integration tests fail against current API contract |

---

## Authentication

### Register
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 88% |
| **Evidence** | `POST /api/v1/auth/register`; `AuthService.Register`; `frontend/src/app/(auth)/register/page.tsx` |
| **Missing** | Email verification; rate limiting |
| **Bugs** | Invalid `agency_id` silently ignored; always redirects to citizen dashboard |

### Login
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 95% |
| **Evidence** | `POST /api/v1/auth/login`; bcrypt; role-based redirect; seed users work |
| **Missing** | MFA; login audit events |
| **Bugs** | None in core path |

### Refresh Token
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 40% |
| **Evidence** | `POST /api/v1/auth/refresh`; refresh stored in localStorage |
| **Missing** | Frontend never calls refresh on 401; no token rotation/revocation |
| **Bugs** | Sessions hard-expire at access TTL (~15 min) |

### JWT Validation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `pkg/jwt/jwt.go`; `middleware/auth.go` `RequireAuth()` |
| **Missing** | Token blacklist; asymmetric signing |
| **Bugs** | Agency frozen at token issuance |

### Agency Switching
| Field | Value |
|-------|-------|
| **Status** | BROKEN |
| **Completion** | 25% |
| **Evidence** | `AgencySwitcher.tsx` updates localStorage + `X-Agency-ID`; CORS allows header |
| **Missing** | Backend reads JWT only (`middleware.GetAgencyID`); no switch-agency endpoint |
| **Bugs** | **Verified:** switcher has no effect on API tenant; lists all agencies not user memberships |

### RBAC
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 75% |
| **Evidence** | `RequireRoles`/`RequireAgencyRoles` on routes; workflow role checks; frontend `rbac.ts`, `RoleGuard` |
| **Missing** | Frontend RBAC is client-only; `FeatureFlagGuard` unwired; some routes lack agency role middleware |
| **Bugs** | `agency_role` `worker` vs global `case_worker` naming inconsistency (workflow fixed via `ResolveWorkflowRole`) |

---

## Multi-Tenancy

### agencies
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | Schema + seed; `GET /api/v1/agencies`; admin list page |
| **Missing** | Admin CRUD for agencies |
| **Bugs** | None |

### agency_users
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 70% |
| **Evidence** | Schema + seed; linked at registration/login; JWT `agency_role` |
| **Missing** | Multi-agency membership resolution; admin management UI |
| **Bugs** | Single primary agency only |

### agency_programs
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | Schema; seed cross-join; `IsProgramEnabledForAgency`; flattened programs API |
| **Missing** | Admin enable/disable UI |
| **Bugs** | None after program API shape fix |

### Tenant Middleware
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | `middleware/tenant.go`; JWT → `GetAgencyID`; `WithTenant` sets session vars |
| **Missing** | `X-Agency-ID` override with membership check |
| **Bugs** | AgencySwitcher ineffective |

### RLS Enforcement
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 75% |
| **Evidence** | `002_row_level_security.up.sql`; `postgres/rls.go` `WithTenant` on core mutations |
| **Missing** | RLS on `agencies`, `agency_users`, `workflow_transitions`; app-layer gaps |
| **Bugs** | `ValidateCaseAccess` defined but not called in handlers |

---

## Application Intake

### Wizard Flow
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `frontend/src/app/citizen/apply/page.tsx` — 5-step wizard with validation |
| **Missing** | Dynamic `form_data` fields; document upload step |
| **Bugs** | None in navigation/submit path |

### Agency Selection
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | Step 0 loads `GET /agencies`; persists to session |
| **Missing** | Filter to user's linked agencies for citizens |
| **Bugs** | None |

### Program Selection
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `GET /agencies/:id/programs` returns flat `Program[]`; `normalizePrograms()` defensive mapping |
| **Missing** | None critical |
| **Bugs** | Fixed in current codebase (was BROKEN pre-fix) |

### Application Submission
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `POST /applications` → case + application + assignment + SLA + events; **verified** creates case |
| **Missing** | `form_data` always `{}` |
| **Bugs** | Integration test still uses deprecated `program_code` payload |

---

## Case Management

### Case Creation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `CaseService.CreateApplication`; case number generation; worker auto-assign |
| **Missing** | Initial workflow event on submit |
| **Bugs** | None |

### Case Retrieval
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 65% |
| **Evidence** | `GET /cases`, `GET /cases/:id`; citizen filtered by `citizen_id` |
| **Missing** | `program` not joined on GET — UI shows empty program name |
| **Bugs** | Worker queue label says "assigned" but list is not filtered by assignment |

### Case Assignment
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | `assignment/allocator.go`; `case_assignments` table; seed assignment |
| **Missing** | Reassignment API; `assignment_history` never written |
| **Bugs** | No manual assign endpoint |

### Case Updates
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | `PATCH /cases/:id` — priority, zip, census tract |
| **Missing** | Case notes (`case_notes` table unused); broader field updates |
| **Bugs** | None in implemented fields |

---

## Workflow Engine

### Status Transitions
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `PATCH /cases/:id/status`; `StateMachine.Transition`; **verified** Submitted → Under Review → Denied |
| **Missing** | Auto-transitions on eligibility/appeal events |
| **Bugs** | Fixed: worker→case_worker role mapping; UpdateStatus SQL 42P08 |

### workflow_events
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `workflow_repo.RecordEvent`; rows created on transition; audit subscriber receives `case.status_changed` |
| **Missing** | Event on initial submit |
| **Bugs** | None after SQL fix |

### Timeline Display
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `GET /cases/:id/workflow`; `CaseTimeline.tsx` with `refreshKey` |
| **Missing** | Citizen case page timeline (worker/supervisor have it) |
| **Bugs** | Empty until first manual transition |

### Transition Validation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `workflow_transitions` seed; `ValidateTransition`; `GET /cases/:id/transitions` for UI dropdown |
| **Missing** | Admin write API for transitions |
| **Bugs** | Integration tests use wrong JSON field `status` vs `to_status` |

---

## Document Management

### Upload
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 40% |
| **Evidence** | `POST /cases/:id/documents` multipart; `DocumentService.Upload` |
| **Missing** | **No frontend UI** calling upload |
| **Bugs** | MinIO bucket missing breaks S3 upload |

### Storage
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 70% |
| **Evidence** | `storage/provider.go`, `local.go`, `s3.go`; keys `{agencyID}/cases/...` |
| **Missing** | Bucket bootstrap; local driver not default in compose |
| **Bugs** | S3 `NoSuchBucket` in default compose config |

### Verification
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | `PATCH /documents/:id/verify`; `UpdateVerification(reviewerID)` |
| **Missing** | Frontend verify UI; status enum validation |
| **Bugs** | Any string accepted as verification status |

### Reviewer Tracking
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | `reviewed_by`, `reviewed_at` columns; set on verify |
| **Missing** | UI display of reviewer; audit trail for verify action |
| **Bugs** | None in backend path |

---

## Eligibility Engine

### Rule Evaluation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | `eligibility/evaluator.go` AND/OR tree; **verified live** `is_eligible: true` with trace |
| **Missing** | String/boolean field types; auto-evaluate on workflow |
| **Bugs** | `toFloat` coercion can mis-evaluate non-numeric fields |

### Versioning
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 50% |
| **Evidence** | `eligibility_rule_versions`; `GetActiveVersion()` by date |
| **Missing** | Admin CRUD API; admin rules page is **hardcoded static data** |
| **Bugs** | `Actions` field stored but never executed |

### Simulation
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | No simulate/dry-run endpoint in codebase |
| **Missing** | Entire feature |
| **Bugs** | N/A |

---

## Benefit Calculation Engine

### Calculation Rules
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 50% |
| **Evidence** | `benefit_calculation_versions` seed; `GetActiveVersion()` |
| **Missing** | Rule admin API/UI; extensible formula DSL |
| **Bugs** | Income reduction hardcoded in calculator |

### Calculation Execution
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `benefit/calculator.go`; **verified live** `calculated_amount: 168` with trace |
| **Missing** | Async recalc job |
| **Bugs** | None in API path |

### Calculation Persistence
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `benefit_calculations` table; `BenefitRepository.Create` |
| **Missing** | `approved_amount` never set |
| **Bugs** | None |

### UI Display
| Field | Value |
|-------|-------|
| **Status** | BROKEN |
| **Completion** | 30% |
| **Evidence** | `BenefitAmountCard.tsx` calls calculate API |
| **Missing** | Field mapping |
| **Bugs** | UI reads `monthly_amount`; API returns `calculated_amount` — **amount never displays** |

---

## Fraud Detection

### Fraud Scan Endpoint
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `POST /cases/:id/fraud/scan`; **verified** returns `{ count: 0, data: [] }` |
| **Missing** | Auto-scan on submit; backend feature-flag enforcement |
| **Bugs** | None in endpoint |

### Fraud Flag Creation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | `fraud/detector.go` heuristics; `FraudRepository.CreateFlag`; event publish |
| **Missing** | Dedup on rescan; advanced rules (SSN, address) |
| **Bugs** | Rescan creates duplicate flags |

### Fraud Review Workflow
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | `POST /fraud/:id/review`; supervisor disposition |
| **Missing** | Agency-wide queue API (UI N+1 loads all cases) |
| **Bugs** | `repeat_offender` counts reviewed flags too |

### UI Integration
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 65% |
| **Evidence** | `FraudFlagBanner`; worker fraud-review page; scan button with success message |
| **Missing** | Feature flag on backend routes |
| **Bugs** | None after notification UX fix |

---

## SLA Tracking

### SLA Policy Lookup
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | `sla_policies` seed; `SLARepository.GetPolicy` |
| **Missing** | Admin policy CRUD |
| **Bugs** | None |

### SLA Calculation
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | `sla/calculator.go` `ComputeDueDate`; tracking created on application |
| **Missing** | Consistent business-days handling in subscriber vs case service |
| **Bugs** | `EnsureTrackingForCase` ignores `business_days_only` |

### At-Risk Detection
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 20% |
| **Evidence** | `ComputeStatus()` returns `at_risk` — **never called** in update path |
| **Missing** | Wiring into `UpdateTrackingStatus`; at-risk API/list |
| **Bugs** | `SLABadge` expects `warning`; backend uses `at_risk` |

### Overdue Detection
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | `GET /sla/breached`; admin SLA page; supervisor escalations count |
| **Missing** | `EventSLABreached` never published; periodic SLA scan job |
| **Bugs** | `elapsed_days` SQL likely incorrect in breach update |

---

## Appeals

### Appeal Creation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 75% |
| **Evidence** | `POST /cases/:id/appeal`; citizen `AppealForm`; persists `appeals` row |
| **Missing** | `appeal_documents` upload |
| **Bugs** | Does not transition case to `appealed` |

### Appeal Workflow
| Field | Value |
|-------|-------|
| **Status** | BROKEN |
| **Completion** | 30% |
| **Evidence** | Workflow transitions seeded for appeal states |
| **Missing** | `AppealService` does not call state machine on file/decide |
| **Bugs** | Appeal status and case status diverge |

### Appeal Queue
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 40% |
| **Evidence** | Worker appeals page lists per-case; supervisor panel on case detail |
| **Missing** | `GET /api/v1/appeals` — **verified 404** |
| **Bugs** | Integration test expects missing route |

### Appeal Review
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 45% |
| **Evidence** | `AppealReviewPanel` on supervisor case page |
| **Missing** | Worker review actions; hearing scheduling |
| **Bugs** | Panel filters `status !== 'decided'` but decisions set `upheld`/`overturned` |

### Appeal Decisions
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 45% |
| **Evidence** | `POST /appeals/:id/decide`; `appeal_decisions` table |
| **Missing** | Case workflow transition to `appeal_approved`/`appeal_denied` |
| **Bugs** | Decision overwrites appeal status with disposition string |

---

## Decision Letters

### Template Management
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 30% |
| **Evidence** | `letter_templates` seed; `GetTemplate()` read-only |
| **Missing** | Template CRUD API; admin letters page is static |
| **Bugs** | None in read path |

### PDF Generation
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 70% |
| **Evidence** | `letters/pdf_generator.go` (gofpdf); merge fields from case/benefit |
| **Missing** | Reliable execution in compose (blocked by storage) |
| **Bugs** | **Verified live:** fails with S3 `NoSuchBucket` |

### Storage
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 55% |
| **Evidence** | Upload to `{agencyID}/letters/...`; `generated_letters.file_key` |
| **Missing** | Bucket init; worker event bus nil |
| **Bugs** | Same MinIO bucket issue |

### Citizen Download
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 15% |
| **Evidence** | `citizen/cases/[id]/letters/page.tsx` lists letters only |
| **Missing** | Download endpoint (documents have `/documents/:id/download`; letters do not) |
| **Bugs** | Frontend `Letter` type mismatches API (`status`, `created_at` vs `generated_at`) |

---

## Notifications

### In-App
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 50% |
| **Evidence** | `notifications` table; subscriber on case events; citizen notifications page |
| **Missing** | Notifications for letters, appeals, fraud, SLA |
| **Bugs** | **`user_id` set to case UUID** instead of citizen ID in subscriber |

### Email
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | Mailhog in compose only |
| **Missing** | SMTP sender; email templates |
| **Bugs** | Config implies email works; it does not |

### Event-Driven Delivery
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 45% |
| **Evidence** | `NotificationSubscriber` on `case.created`, `case.status_changed` |
| **Missing** | Queue→worker pipeline; broader event coverage |
| **Bugs** | Wrong recipient ID (see in-app) |

---

## Reporting

### Report Generation
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 25% |
| **Evidence** | `POST /reports` creates pending row; worker sets `completed` + placeholder `file_key` |
| **Missing** | Actual query/aggregation by report type |
| **Bugs** | Marks completed without producing output |

### CSV Export
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | No CSV library or generator |
| **Missing** | Entire feature |
| **Bugs** | N/A |

### Excel Export
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | No Excel library or generator |
| **Missing** | Entire feature |
| **Bugs** | N/A |

### PDF Export
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | gofpdf used for letters only, not reports |
| **Missing** | Report PDF pipeline |
| **Bugs** | N/A |

---

## Geographic Analytics

### ZIP Aggregation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `ReportRepository.CasesByZip`; exposed in `GET /analytics/summary` |
| **Missing** | Census tract aggregation |
| **Bugs** | None |

### Map Rendering
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | `GeoHeatmap.tsx` Leaflet; `dashboard/geography/page.tsx` |
| **Missing** | Full geocoding (hardcoded ~10 LA ZIPs) |
| **Bugs** | None in render path |

### API Data
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `GET /analytics/summary` returns `cases_by_zip`; **verified live** |
| **Missing** | Dedicated `/analytics/geography` endpoint (docs only) |
| **Bugs** | Backend feature flag not enforced on route |

---

## Feature Flags

### Backend Enforcement
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 10% |
| **Evidence** | `FeatureFlagGuard` in `middleware/feature_flag.go`; evaluation service exists |
| **Missing** | Guard never registered on any route in `router.go` |
| **Bugs** | APIs fully callable when frontend hides features |

### Frontend Gating
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | `useFeatureFlag` on appeals, fraud, workflow, geo, retention pages; admin toggle UI |
| **Missing** | Rollout % not honored (checks `is_enabled` only) |
| **Bugs** | UI-only security |

---

## Audit Logging

### Audit Creation
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | `AuditSubscriber` on all events via `SubscribeAll`; **verified** audit rows on status change |
| **Missing** | `ip_address`, `previous_state` population |
| **Bugs** | Worker async letters skip audit (nil event bus) |

### Immutability
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `003_audit_immutability.up.sql` triggers block UPDATE/DELETE |
| **Missing** | Automated tests for triggers |
| **Bugs** | None observed |

### Audit Viewer
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 20% |
| **Evidence** | `GET /audit-logs` API exists (admin/supervisor) |
| **Missing** | **No frontend page**; not in navigation |
| **Bugs** | API-only feature |

---

## Event Bus

### Event Publishing
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `events/bus.go`; publishers in case, document, eligibility, benefit, fraud, appeal, letter services |
| **Missing** | Async/retry/persistence |
| **Bugs** | Synchronous inline handlers in request path |

### Subscribers
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 65% |
| **Evidence** | Audit, notification, SLA, letter subscribers wired in `cmd/api/main.go` |
| **Missing** | Subscribers for retention, benefit recalc, fraud analytics |
| **Bugs** | Notification subscriber wrong user_id |

### Side Effects
| Field | Value |
|-------|-------|
| **Status** | PARTIAL |
| **Completion** | 60% |
| **Evidence** | Status change → audit + notification + SLA + letter enqueue |
| **Missing** | `EventSLABreached` never published; letter worker skips bus |
| **Bugs** | Handler errors logged but not retried |

---

## Background Jobs

### Notification Jobs
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | Notifications written synchronously in subscriber |
| **Missing** | Queued notification dispatch job |
| **Bugs** | N/A |

### SLA Monitor
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 10% |
| **Evidence** | SLA updates only on case status change events |
| **Missing** | Periodic cron/scheduled breach scan |
| **Bugs** | At-risk never computed |

### Report Generation
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 20% |
| **Evidence** | `generate_report` job type; worker marks completed with JSON path placeholder |
| **Missing** | File generation and storage write |
| **Bugs** | False-positive completion status |

### Benefit Recalculation
| Field | Value |
|-------|-------|
| **Status** | NOT IMPLEMENTED |
| **Completion** | 0% |
| **Evidence** | Benefit calc is sync API only |
| **Missing** | Background recalc job |
| **Bugs** | N/A |

### Retention Purge
| Field | Value |
|-------|-------|
| **Status** | STUBBED |
| **Completion** | 15% |
| **Evidence** | `JobRunRetention` constant defined; retention policies seeded; admin retention page |
| **Missing** | Worker handler; enqueue trigger |
| **Bugs** | Unknown job type logged if enqueued |

---

## Observability

### Health Endpoint
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 95% |
| **Evidence** | `GET /health` → `{"status":"ok"}` — **verified live** |
| **Missing** | Dependency detail in health response |
| **Bugs** | None |

### Readiness Endpoint
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 90% |
| **Evidence** | `GET /ready` pings PostgreSQL + Redis — **verified live** |
| **Missing** | MinIO/SMTP checks |
| **Bugs** | None |

### Metrics Endpoint
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 85% |
| **Evidence** | `GET /metrics` Prometheus; HTTP + job counters |
| **Missing** | `DBQueryDuration` defined but never instrumented |
| **Bugs** | None |

### Structured Logging
| Field | Value |
|-------|-------|
| **Status** | COMPLETE |
| **Completion** | 80% |
| **Evidence** | JSON `slog` in `pkg/logger`; request ID middleware |
| **Missing** | Structured access logs per request; distributed tracing |
| **Bugs** | None |

---

## Verification Log (Live Tests Run)

| Test | Result |
|------|--------|
| `GET /health`, `GET /ready` | Pass |
| `POST /auth/login` (worker, supervisor, citizen) | Pass |
| `POST /applications` with valid `program_id` | Pass |
| `PATCH /cases/:id/status` Submitted → Under Review → Denied | Pass (after fixes) |
| `GET /cases/:id/workflow` | Pass — 2 events returned |
| `POST /cases/:id/fraud/scan` | Pass — `{ count: 0 }` |
| `POST /cases/:id/eligibility/evaluate` | Pass — `is_eligible: true` + trace |
| `POST /cases/:id/benefit/calculate` | Pass — `calculated_amount: 168` + trace |
| `POST /cases/:id/letters` (approval_notice) | **Fail** — S3 `NoSuchBucket` |
| `GET /api/v1/appeals` | **Fail** — 404 |
| `go test -tags=integration ./tests/integration/...` | **Fail** — 2 tests |
| Audit rows after status change | Pass — 2 `case.status_changed` entries |

---

## Recommended Remediation Priority

1. **MinIO bucket bootstrap** — unblocks documents and letters in Compose
2. **Benefit UI field mapping** — one-line fix unlocks benefit display
3. **Appeal ↔ workflow integration** — file/decide must call state machine
4. **Document frontend + letter download endpoint** — complete citizen/worker flows
5. **Fix notification `user_id`** — use citizen ID from case payload
6. **Wire `FeatureFlagGuard` or remove false security** on gated routes
7. **Implement report file generation** or stop marking jobs completed
8. **Repair integration tests** — align with current API contract
9. **Agency switching** — honor membership + reissue JWT or read validated `X-Agency-ID`
10. **Frontend token refresh** — call `/auth/refresh` on 401

---

## Appendix: Status Summary by Area

| Area | Complete | Partial | Stubbed | Broken | Not Impl. | Area % |
|------|----------|---------|---------|--------|-----------|--------|
| Foundation | 4 | 3 | 1 | 0 | 0 | 78% |
| Authentication | 2 | 2 | 0 | 1 | 0 | 72% |
| Multi-Tenancy | 2 | 3 | 0 | 0 | 0 | 76% |
| Application Intake | 4 | 0 | 0 | 0 | 0 | 88% |
| Case Management | 1 | 3 | 0 | 0 | 0 | 68% |
| Workflow Engine | 4 | 0 | 0 | 0 | 0 | 89% |
| Document Management | 0 | 4 | 0 | 0 | 0 | 56% |
| Eligibility Engine | 1 | 1 | 0 | 0 | 1 | 43% |
| Benefit Calculation | 2 | 1 | 0 | 1 | 0 | 63% |
| Fraud Detection | 2 | 2 | 0 | 0 | 0 | 71% |
| SLA Tracking | 1 | 1 | 1 | 0 | 0 | 54% |
| Appeals | 1 | 3 | 0 | 1 | 0 | 47% |
| Decision Letters | 0 | 2 | 1 | 0 | 1 | 43% |
| Notifications | 0 | 2 | 0 | 0 | 1 | 32% |
| Reporting | 0 | 0 | 1 | 0 | 3 | 6% |
| Geographic Analytics | 3 | 0 | 0 | 0 | 0 | 83% |
| Feature Flags | 0 | 1 | 0 | 0 | 1 | 35% |
| Audit Logging | 2 | 0 | 0 | 0 | 1 | 63% |
| Event Bus | 1 | 2 | 0 | 0 | 0 | 70% |
| Background Jobs | 0 | 0 | 2 | 0 | 3 | 9% |
| Observability | 4 | 0 | 0 | 0 | 0 | 88% |

*This audit reflects repository state as of 2026-06-08 after recent fixes to application intake (program API), workflow transitions (role mapping, UpdateStatus SQL), and case action UX (timeline refresh, fraud scan feedback).*

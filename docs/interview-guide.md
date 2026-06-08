# Interview Guide — Government Benefits Platform

Use this guide to prepare for SWE, government/public-sector, and backend-focused interviews. Answers reference concrete decisions in this repository.

---

## General / System Design

### "Walk me through this project."

**Answer:** I built a multi-tenant government benefits platform modeled after LA County-style public assistance. Citizens apply through a Next.js wizard; cases move through a configurable workflow (submitted → review → eligibility → supervisor decision). Workers evaluate JSON eligibility rules and calculate benefits; supervisors approve or deny and decide appeals. Every significant action publishes a domain event that writes to an immutable audit log. The stack is Go/Gin, PostgreSQL with RLS, Redis for async jobs, MinIO for documents and PDF letters, and Docker Compose for deployment. I verified the full path with integration tests and a Playwright demo script.

### "Why Go for the backend?"

**Answer:** Go fits IO-heavy API workloads with strong concurrency, static binaries for containers, and clear boundaries for services and repositories. For a portfolio focused on public-sector backend roles, Go also maps well to common agency contractor stacks. The API uses Gin for routing/middleware, pgx for PostgreSQL, and table-driven tests for workflow logic.

### "How would you scale this?"

**Answer:** Horizontally scale stateless API instances behind a load balancer; use managed PostgreSQL with read replicas for reporting; move Redis to a managed cluster; replace MinIO with S3 in production. The worker process already handles async letter generation separately from request latency. Audit and analytics queries can move to read replicas or a warehouse without touching the OLTP path.

---

## Government / Public Sector Engineers

### "How do you handle PII and compliance?"

**Answer:** Tenant isolation uses agency-scoped JWT claims, middleware validation, repository filters, and PostgreSQL RLS. Audit logs are append-only with a database trigger blocking deletes. SSN is stored hashed in user profiles; documents live in object storage with metadata in PostgreSQL. Role-based access separates citizen, worker, supervisor, and admin capabilities. The design demonstrates patterns agencies expect even though this is a demo—not a certified production system.

### "How does your workflow model map to real case processing?"

**Answer:** Each agency defines allowed transitions in `workflow_transitions` with a required role. Statuses mirror common public assistance stages: intake, document requests, eligibility review, supervisor review, approval/denial, appeals, and closure. Every transition creates a `workflow_event` for the case timeline and an audit entry. Supervisors can approve or deny at review; citizens can appeal denials; supervisors decide appeals with a recorded rationale.

### "What would you change before a pilot with a county?"

**Answer:** Security review (pen test, SB 272-style privacy assessment), SSO integration (SAML/OIDC), formal data retention enforcement, WCAG audit beyond baseline styling, agency switching hardening, production SMTP and monitoring, and operational runbooks. I'd also add rate limiting, WAF rules, and secrets management via a vault rather than flat `.env` files.

---

## Backend Engineers

### "Explain multi-tenancy in your database."

**Answer:** Every tenant-scoped table includes `agency_id`. On each request, middleware resolves the active agency from JWT + `X-Agency-ID`, sets it on the request context, and repositories filter queries accordingly. PostgreSQL RLS policies add a second layer: even a miswritten query cannot cross agencies if the session variable is set correctly. This defense-in-depth pattern is important when multiple agencies share one database cluster.

### "How does the eligibility engine work?"

**Answer:** Rules are versioned JSON condition trees per program—comparisons on household size, income, employment status, etc. The evaluator walks the tree, records a trace of each condition pass/fail, and persists the result on `eligibility_evaluations`. Workers trigger evaluation via API; the UI displays eligibility status and trace for caseworker review. New rule versions can be effective-dated without code deploys.

### "How do you prevent duplicate appeal decisions?"

**Answer:** `appeal_decisions` has a unique constraint on `appeal_id`. The service checks for an existing decision before insert and returns a business-friendly 409 conflict. The pending appeals API filters to filed appeals without a decision row and cases in appealable statuses. The UI disables action buttons after submission and refreshes the queue.

### "Describe your event bus."

**Answer:** An in-process pub/sub bus publishes typed domain events after successful transactions. Subscribers handle audit logging, notifications, SLA updates, and letter job enqueueing. This keeps handlers thin and side effects decoupled. In production I'd consider an outbox pattern or message broker for durability across API restarts.

---

## Frontend / Full-Stack

### "How is authorization handled in the UI?"

**Answer:** JWT stored client-side with refresh support; `resolvePrimaryRole` maps user roles to citizen/worker/supervisor/admin. Route guards in layout components block unauthorized paths. Navigation items come from a central RBAC map. API calls attach Bearer token and agency header. The UI fails gracefully with user-readable errors instead of raw HTTP statuses.

### "How did you verify the demo workflow?"

**Answer:** Three layers: Go integration tests for apply → approve and deny → appeal paths; a PowerShell script with 16 API verification steps; and a Playwright headed recording spec that follows the written demo script with reset/seed scripts for reproducibility.

---

## Behavioral / Process

### "Tell me about a hard bug you fixed."

**Answer:** Decided appeals still appeared in the supervisor pending queue because the list API returned all filed appeals regardless of decision state, and duplicate clicks hit a unique constraint with a raw SQL error. I added server-side pending filters, pre-checks before insert, mapped unique violations to 409 responses with clear messages, and updated the frontend to use the pending endpoint and disable buttons after decision. Verified with automated tests from a clean database.

### "What would you do differently if you started over?"

**Answer:** Generate OpenAPI clients for the frontend early, use an outbox for events from day one, and keep sample transactional data out of migration seeds—only reference data in migrations, demo data in separate seed scripts (which we added later).

---

## Quick Reference — Numbers to Cite

| Metric | Value |
|--------|------:|
| REST endpoints | 40+ |
| Database tables | 25+ |
| Role portals | 4 |
| Workflow statuses | 15+ |
| Docker services | 6 |
| Integration test path | Apply → eligibility → benefit → approve/deny → appeal → audit |

Demo accounts and password: see [README.md](../README.md)

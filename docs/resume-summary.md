# Resume Summary — Government Benefits Platform

Tailor this document for LA County, California state agencies, public-sector contractors, and enterprise backend roles.

---

## Project Summary

Full-stack **Government Benefits & Case Management Platform** — a multi-tenant public assistance system covering intake, eligibility rules, benefit calculation, workflow-driven case processing, appeals, PDF correspondence, and immutable audit compliance.

Built with **Go**, **PostgreSQL (RLS)**, **Redis**, **MinIO**, and **Next.js 15**. Designed to demonstrate patterns used in county and state benefits administration: role-based portals, configurable workflows, rules engines, and compliance-grade audit trails.

**Scope:** 40+ REST endpoints · 25+ database tables · 4 role portals · 15+ workflow statuses · integration-tested apply-to-appeal path · Docker Compose deployment · GitHub Actions CI

---

## Architecture Highlights

- **Multi-tenant isolation** — PostgreSQL row-level security with agency-scoped JWT claims and validated `X-Agency-ID` header switching
- **Event-driven audit** — Domain event bus with subscribers for audit logs, notifications, SLA tracking, and async letter jobs
- **Workflow state machine** — Configurable transitions with role-based guards (citizen, worker, supervisor)
- **Rules engine** — JSON-defined eligibility conditions and benefit formulas with versioned effective dates
- **Async processing** — Redis-backed worker for PDF letter generation and report jobs
- **Production ops** — Docker Compose (dev + prod), health/readiness probes, GitHub Actions CI with integration tests

---

## Engineering Challenges Solved

| Challenge | Approach |
|-----------|----------|
| **Tenant data isolation in a shared database** | Combined JWT claims, middleware validation, repository filters, and PostgreSQL RLS policies |
| **Compliance-grade audit trail** | Immutable `audit_logs` table with DB triggers; event bus subscribers capture actor, entity, and state diffs |
| **Complex case lifecycle** | Agency-configurable workflow transitions with role guards; timeline built from `workflow_events` |
| **Rules without redeploying code** | Versioned eligibility and benefit rule JSON stored in PostgreSQL; evaluator interprets conditions at runtime |
| **Async PDF generation** | Letter jobs enqueued to Redis; worker generates PDFs via gofpdf and stores in MinIO |
| **Appeal integrity** | Unique constraint on appeal decisions; pending-queue filtering; friendly conflict responses on duplicate decisions |
| **Demo reliability** | Reset/seed SQL scripts, API verification suite, Playwright recording spec for repeatable walkthroughs |

---

## Recommended Resume Bullets

Pick 3–5 depending on role emphasis.

**Backend / Platform**

- Designed and implemented a **multi-tenant Go REST API** with PostgreSQL row-level security, JWT authentication with refresh tokens, and agency-scoped tenant middleware for a government benefits case management platform serving 40+ endpoints.

- Built an **event-driven audit subsystem** publishing domain events (application, eligibility, benefit, appeal, letter) to an immutable audit log with actor attribution, search, and pagination for compliance review.

- Implemented a **configurable workflow state machine** with role-based transition guards, SLA tracking, and worker assignment for public assistance case lifecycle management.

**Full-Stack / Product**

- Delivered a **role-based Next.js portal suite** (citizen intake wizard, worker case review, supervisor approvals, admin rules/audit) integrated with a Go backend and real-time eligibility/benefit evaluation UI.

- Shipped **end-to-end benefits workflow** from citizen application through eligibility determination, benefit calculation, supervisor decision, PDF letter generation, and appeals — verified by automated integration and Playwright tests.

**DevOps / Reliability**

- Containerized the platform with **Docker Compose** (development and production overlays), readiness checks for PostgreSQL/Redis/MinIO, and GitHub Actions CI running lint, unit tests, integration tests, and multi-image builds.

- Hardened production release with **token refresh flows**, file upload validation, user-facing error messaging, and `.env`-driven secret management for deployable release candidates.

**Public Sector / Domain**

- Modeled **LA County-style public assistance programs** (food assistance, housing, emergency relief, healthcare) with versioned eligibility rules, benefit formulas, denial/approval correspondence, and citizen appeal workflows suitable for government agency demonstrations.

---

## Keywords for ATS

Go, Golang, PostgreSQL, Redis, REST API, JWT, multi-tenant, row-level security, microservices, Docker, CI/CD, GitHub Actions, Next.js, TypeScript, React, Tailwind, government, public sector, case management, workflow engine, audit trail, compliance, S3, MinIO, event-driven architecture

---

## Links to Include on Resume

- GitHub: https://github.com/yucheng1554439/gov-benefits-platform
- Live demo URL (when deployed)
- Supporting docs: `README.md`, `docs/demo-readiness-report.md`, `docs/interview-guide.md`

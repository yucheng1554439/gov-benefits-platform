# System Architecture

## Overview

The Government Benefits & Case Management Platform is a multi-tenant SaaS system for public assistance program intake, case management, eligibility determination, benefit calculation, and appeals.

## Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| Frontend | Next.js 15, TypeScript, Tailwind | Citizen portal, worker/supervisor/admin UIs |
| API | Go 1.25, Gin | REST API with JWT auth and RBAC |
| Worker | Go background process | Async jobs via Redis queue |
| Database | PostgreSQL 16 + RLS | Primary data store with tenant isolation |
| Cache/Queue | Redis 7 | Job queue, feature flag cache |
| Storage | MinIO/S3 | Document and letter PDF storage |
| Email | Mailhog (dev) / SMTP (prod) | Notification delivery |

## Architecture Diagram

```
Citizens / Workers / Admins
         │
         ▼
   Next.js Frontend (port 3000)
         │
         ▼
   Go Gin API (port 8080)
    ├── Tenant Middleware (agency_id)
    ├── RBAC Middleware
    ├── Feature Flag Gating
    └── Event Bus
         ├── Audit Subscriber
         ├── Notification Subscriber → Redis Queue → Worker
         ├── SLA Subscriber
         └── Letter Subscriber → Redis Queue → Worker
         │
         ▼
   PostgreSQL (RLS) + MinIO + Redis
```

## Multi-Tenancy

Every tenant-scoped record includes `agency_id`. Isolation enforced at:
1. JWT claims (`agency_id`, `agency_ids[]`)
2. API middleware (`X-Agency-ID` header)
3. Repository queries (`WHERE agency_id = ?`)
4. PostgreSQL Row-Level Security policies

## Event-Driven Design

Domain events (`case.approved`, `appeal.filed`, etc.) are published to an in-process event bus. Subscribers handle side effects without direct service-to-service coupling.

## Security

- JWT authentication with refresh tokens
- bcrypt password hashing (cost 12)
- Parameterized SQL queries (pgx)
- Immutable audit logs (DB triggers)
- CORS restricted to frontend origin
- Trivy security scanning in CI

## Deployment

```bash
docker compose -f infra/compose/docker-compose.yml up -d --build
```

Production: horizontally scaled API behind load balancer, managed PostgreSQL, S3 storage, secrets via vault.

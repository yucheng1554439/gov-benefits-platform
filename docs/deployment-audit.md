# Deployment Audit

**Date:** 2026-06-08  
**Frontend:** Vercel — https://gov-benefits-platform.vercel.app  
**Backend:** Render — https://gov-benefits-platform.onrender.com

---

## Production URLs

| Service | URL | Health check |
|---------|-----|--------------|
| Frontend | https://gov-benefits-platform.vercel.app | Login page loads (200) |
| Backend API | https://gov-benefits-platform.onrender.com/api/v1 | `GET /health` → `{"status":"ok"}` |
| Backend root | https://gov-benefits-platform.onrender.com | Redirects to API |

---

## Vercel (Frontend)

| Variable | Expected value | Status |
|----------|----------------|--------|
| `NEXT_PUBLIC_API_URL` | `https://gov-benefits-platform.onrender.com/api/v1` | **Verify in Vercel dashboard** — frontend loads and login reaches Render API |

### Verification

- [x] https://gov-benefits-platform.vercel.app/login renders Sign In form
- [x] Demo account hints visible
- [ ] Confirm `NEXT_PUBLIC_API_URL` in Vercel → Settings → Environment Variables matches table above (dashboard access required)

### Notes

- Vercel builds from `frontend/` (Next.js standalone).
- After backend URL changes, trigger a redeploy so build-time `NEXT_PUBLIC_*` is baked in.

---

## Render (Backend)

| Variable | Expected value | Status |
|----------|----------------|--------|
| `DATABASE_URL` | PostgreSQL connection string (Render Postgres or external) | **Required** — API `/health` OK; `/api/v1/agencies` returns seed agencies |
| `REDIS_URL` | Redis connection string | **Required** for async jobs — verify in Render dashboard |
| `JWT_SECRET` | Strong random secret (not `change-me-in-production`) | **Required** — login works in production |
| `STORAGE_DRIVER` | `local` or `s3` | **Required** — use `local` on Render disk or `s3` with MinIO/S3 vars |
| `LOCAL_STORAGE_PATH` | e.g. `/var/data/govbenefits` when `STORAGE_DRIVER=local` | Set if using local storage |
| `CORS_ORIGIN` | `https://gov-benefits-platform.vercel.app` | **Critical** — must match Vercel origin exactly (no trailing slash) |

### Additional backend vars (from `.env.example`)

| Variable | Production guidance |
|----------|---------------------|
| `PORT` | Render sets automatically (`8080` typical) |
| `ENVIRONMENT` | `production` |
| `JWT_ACCESS_TTL` / `JWT_REFRESH_TTL` | Defaults OK (900 / 604800) |
| `S3_*` | Required only if `STORAGE_DRIVER=s3` |
| `SMTP_*` | Optional — notifications not required for demo |

### Verification (API, 2026-06-08)

| Check | Result |
|-------|--------|
| `GET /health` | **PASS** |
| `GET /api/v1/agencies` | **PASS** (4 agencies) |
| `POST /api/v1/auth/login` (worker) | **PASS** |
| `GET /api/v1/cases` | **PASS** (returns cases; program names pending backend redeploy) |
| CORS from browser | **PASS** (login works from Vercel origin) |

---

## Cross-service checklist

| Item | Expected |
|------|----------|
| Frontend API base | Points to Render `/api/v1` |
| Backend CORS | Allows Vercel frontend origin |
| Seed data | Migrations run on deploy; demo users exist |
| HTTPS | Both services on HTTPS |

---

## Recommended Render env block (copy/paste template)

```env
ENVIRONMENT=production
DATABASE_URL=<from Render Postgres>
REDIS_URL=<from Render Redis or Upstash>
JWT_SECRET=<openssl rand -hex 32>
STORAGE_DRIVER=local
LOCAL_STORAGE_PATH=/var/data/govbenefits
CORS_ORIGIN=https://gov-benefits-platform.vercel.app
```

---

## Post-deploy actions

1. Redeploy Render after program-list fix (`backend/internal/service/case.go`).
2. Confirm Vercel `NEXT_PUBLIC_API_URL` unchanged after backend redeploy.
3. Run `docs/final-release-report.md` verification checklist in browser.

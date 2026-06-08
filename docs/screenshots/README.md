# Portfolio Screenshots

Capture these screens after running `scripts/seed_demo_data.ps1` (or completing the live demo flow). Save PNG files in this directory using the filenames below so the root README renders correctly.

| File | Route / action | Login |
|------|----------------|-------|
| `citizen-apply.png` | `/citizen/apply` — Food Assistance wizard | citizen1@example.com |
| `worker-review.png` | `/worker/cases/{id}` — eligibility + benefit cards | worker1@dpss.lacounty.gov |
| `benefit-calculation.png` | Worker case — Benefit Amount card after Calculate | worker1@dpss.lacounty.gov |
| `appeals-review.png` | `/supervisor/appeals` — pending appeal panel | supervisor1@dpss.lacounty.gov |
| `audit-trail.png` | `/admin/audit` — filtered event list | admin@dpss.lacounty.gov |
| `analytics-dashboard.png` | `/dashboard/analytics` | supervisor1@dpss.lacounty.gov |

Password for all accounts: `Password123!`

## Automated capture

```powershell
docker compose -f infra/compose/docker-compose.yml up -d
.\scripts\seed_demo_data.ps1
npx playwright test tests/e2e/portfolio-screenshots.spec.ts --project=portfolio-screenshots
```

# 5-Minute Demo Script

Exact UI steps for a recorded demonstration. Password for all accounts: `Password123!`

**Before recording:** run `scripts/seed_demo_data.ps1` (or `.sh`) for curated showcase cases, **or** `scripts/reset_demo_data.ps1` for a completely live walkthrough.

---

## Option A — Live Walkthrough (recommended for interviews)

### 1. Citizen applies

1. Open `http://localhost:3000/login`
2. Sign in: `citizen1@example.com` / `Password123!`
3. Click **Apply** in the sidebar
4. Complete the application wizard:
   - Agency: LA County DPSS
   - Program: **Food Assistance**
   - Household size: `3`
   - Annual income: `$24,000`
   - Submit
5. On the dashboard, open the new case (case number `CASE-2026-…`)

### 2. Worker reviews

1. Sign out → sign in: `worker1@dpss.lacounty.gov`
2. Click **Queue** in the sidebar
3. Open the submitted case
4. Click **Fraud Scan**
5. In the right column, click **Evaluate Eligibility** on the Eligibility card
6. Click **Calculate Benefit** on the Benefit Amount card
7. Under **Update Status**, select **Under Review** → click **Update**
8. Select **Eligibility Review** → click **Update**

### 3. Supervisor approves

1. Sign out → sign in: `supervisor1@dpss.lacounty.gov`
2. Click **Escalations**
3. Open the case from the table
4. Click **Approve** (or **Deny** if demonstrating appeals)

### 4. Worker generates letter

1. Sign in as `worker1@dpss.lacounty.gov`
2. Open the approved case
3. In the **Correspondence** card, click **Generate Approval Letter**
4. Sign in as `citizen1@example.com`
5. Open the same case → **View Letters** → **Download**

### 5. Appeal path (if case was denied)

1. Sign in as `citizen1@example.com`
2. Open denied case → **File Appeal**
3. Enter grounds → **File Appeal**

### 6. Supervisor decides appeal

1. Sign in as `supervisor1@dpss.lacounty.gov`
2. Click **Appeals** in the sidebar (not Worker Appeals)
3. Select the pending appeal
4. Enter rationale
5. Click **Approve Appeal** or **Deny Appeal**
6. Confirm success message; appeal disappears from pending list

### 7. Admin audit

1. Sign in as `admin@dpss.lacounty.gov`
2. Click **Audit Trail**
3. Search or filter by `appeal.decided`, `benefit.calculated`, etc.

---

## Option B — Curated Dataset (fast portfolio tour)

After `scripts/seed_demo_data.ps1`:

| Case | Number | Program | Status | Show |
|------|--------|---------|--------|------|
| A | CASE-2026-DEMO-A | Food Assistance | Approved | Eligibility, benefit, letter, audit |
| B | CASE-2026-DEMO-B | Housing Assistance | Appealed | Pending appeal → Supervisor → Appeals |
| C | CASE-2026-DEMO-C | Emergency Relief | Under Review | Worker → Fraud Review |
| D | CASE-2026-DEMO-D | Healthcare Assistance | Need Documents | Document workflow state |

1. Worker **Queue** → open DEMO-A → show benefit/eligibility data
2. Supervisor **Appeals** → decide DEMO-B live
3. Worker **Fraud Review** → show DEMO-C flag
4. Worker **Queue** → open DEMO-D → show Need Documents status
5. Admin **Audit Trail** → filter events

---

## Accounts

| Role | Email |
|------|-------|
| Citizen | citizen1@example.com |
| Worker | worker1@dpss.lacounty.gov |
| Supervisor | supervisor1@dpss.lacounty.gov |
| Admin | admin@dpss.lacounty.gov |

---

## Reset between takes

```powershell
.\scripts\reset_demo_data.ps1
.\scripts\seed_demo_data.ps1   # optional curated cases
```

Automated check:

```powershell
.\scripts\run_demo_verification.ps1
```

### Automated screen recording (Playwright)

With Docker services running on `localhost:3000`:

```powershell
npm run demo:install   # first time only
npm run demo           # reset DB + headed browser walkthrough (~2 min)
```

Tune pacing with `DEMO_PAUSE_MS`, `DEMO_SLOW_MO`, and `DEMO_ACTION_MS`.

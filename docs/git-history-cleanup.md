# Git History Cleanup — Remove Cursor Co-Author

Goal: single clean commit on `main` authored only by your GitHub account, with **no file loss**.

> **Warning:** This rewrites `main` history. Coordinate with anyone else using the repo.  
> **Do not run** until you have pushed any other local work and confirmed the working tree is clean.

---

## Exact commands (run from repo root)

Replace `Your Name` and `your@email.com` with your git identity (must match GitHub noreply if you use that).

### 1. Safety backup

```powershell
cd C:\Users\28534\Projects\gov-benefits-platform

git status
git branch backup-before-history-rewrite

git push origin backup-before-history-rewrite
```

### 2. Create orphan branch with current tree

```powershell
git checkout --orphan portfolio-release-clean

git add -A
git status
```

Confirm **all** project files are staged (backend, frontend, docs, scripts, tests, infra, `.github`, etc.).

### 3. Single clean commit (no Co-authored-by trailer)

```powershell
git commit -m "Portfolio release: Government Benefits Platform" -m "Production-ready demo with citizen intake, case workflow, appeals, audit trail, and deployed Vercel/Render stack."
```

Verify the commit message has **no** `Co-authored-by:` line:

```powershell
git log -1 --format=full
```

### 4. Replace `main`

```powershell
git branch -M main
```

### 5. Force-push (required after history rewrite)

```powershell
git push --force origin main
```

### 6. Verify on GitHub

- Open https://github.com/yucheng1554439/gov-benefits-platform/commits/main
- Confirm **one** (or your chosen) commit(s) with only your author
- Confirm CI badges still pass
- Confirm Vercel/Render redeploy from the new commit

---

## Optional: set author explicitly for the orphan commit

If your global git config is not your GitHub identity:

```powershell
git -c user.name="Your Name" -c user.email="your@email.com" commit -m "Portfolio release: Government Benefits Platform"
```

---

## Rollback if needed

```powershell
git fetch origin
git checkout main
git reset --hard origin/backup-before-history-rewrite
git push --force origin main
```

---

## What this does **not** change

- File contents (identical to pre-rewrite tree)
- Remote backup branch preserves old history
- GitHub Actions, Vercel, and Render configs remain in the repo

## What this **does** change

- All prior commits (including `Co-authored-by: Cursor`) disappear from `main` history
- Commit SHAs change — update any pinned links if needed

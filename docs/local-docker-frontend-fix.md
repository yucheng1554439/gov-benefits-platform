# Local Docker Frontend Fix

## Symptom

The frontend container exited immediately on startup:

```text
Error: Cannot find module '/app/server.js'
```

`docker compose logs frontend` showed Node failing because the production entrypoint did not exist at the path the Dockerfile expected.

## Root cause

The frontend uses Next.js **standalone** output (`output: 'standalone'` in `next.config.ts`). In a Docker multi-stage build, the runner stage copies `.next/standalone` to `/app` and runs `node server.js`.

`next.config.ts` also set `outputFileTracingRoot` to the **repository root** (`path.join(__dirname, '..')`) so local Playwright tests can resolve dependencies from the monorepo lockfile. That setting changes how Next.js nests the standalone bundle:

| Build context | Standalone server path |
|---------------|------------------------|
| Local (with `outputFileTracingRoot`) | `.next/standalone/frontend/server.js` |
| Docker (WORKDIR `/app`, same setting) | `.next/standalone/app/server.js` |

The Dockerfile copied the entire `.next/standalone` tree to `/app` but ran `CMD ["node", "server.js"]`, which only works when `server.js` sits **directly** under the standalone root (`.next/standalone/server.js`).

With `outputFileTracingRoot` enabled in the image build, `server.js` was nested one level deeper, so the container could not start.

## Fix

Follow Next.js 15 production deployment for standalone Docker images:

1. **Disable monorepo tracing root during Docker builds** — set `DOCKER_BUILD=1` in the builder stage so standalone output is flat (`server.js` at `.next/standalone/server.js`).
2. **Keep `outputFileTracingRoot` for local dev/Playwright** — only apply it when `DOCKER_BUILD !== '1'`.
3. **Ensure `public/` exists in the build context** — add `frontend/public/.gitkeep` so `COPY public` in the runner stage succeeds even when there are no static assets yet.

No CMD or copy-path workarounds (e.g. `node app/server.js` or copying a nested subdirectory) were used; the image matches the documented standalone layout.

## Files changed

| File | Change |
|------|--------|
| `frontend/next.config.ts` | Apply `outputFileTracingRoot` only when `DOCKER_BUILD !== '1'`. |
| `infra/docker/Dockerfile.frontend` | Set `ENV DOCKER_BUILD=1` before `npm run build`. |
| `frontend/public/.gitkeep` | Placeholder so the `public` directory exists for Docker `COPY`. |

## Verification

### Build and standalone layout

```powershell
docker compose -f infra/compose/docker-compose.yml build frontend
```

Inside a one-off container from the built image, `/app/server.js` exists and is the correct entrypoint.

### Compose startup

```powershell
docker compose -f infra/compose/docker-compose.yml up -d --build frontend
docker compose -f infra/compose/docker-compose.yml ps frontend
docker compose -f infra/compose/docker-compose.yml logs frontend
```

**Result:** `compose-frontend-1` stays **Up**; logs show:

```text
▲ Next.js 15.5.19
- Local:        http://localhost:3000
✓ Ready in 76ms
```

### HTTP checks

| URL | Status |
|-----|--------|
| http://localhost:3000/ | 200 |
| http://localhost:3000/login | 200 |

## References

- [Next.js standalone output](https://nextjs.org/docs/app/api-reference/config/next-config-js/output#automatically-copying-traced-files)
- [Next.js Docker example (standalone)](https://nextjs.org/docs/app/building-your-application/deploying#docker-image)

# API deployment and rollback

The API is deployed as an immutable GHCR image tagged with the source commit SHA.
The `Deploy API` workflow runs after a successful `CI` workflow on `main`. Configure
GitHub `staging` and `production` environments with the same variable and secret
names below, but different values. Add required reviewers to `production` so a
successful build does not bypass a human production approval. Manual runs default
to staging; successful CI on `main` creates a pending production deployment.

## Required environment configuration

For each environment, set `API_PUBLIC_URL` to its public API origin, without a
trailing slash. Add these environment secrets:

- `VPS_HOST`, `VPS_PORT`, `VPS_USER`
- `VPS_SSH_PRIVATE_KEY`
- `VPS_SSH_KNOWN_HOSTS` — pin the server host key; do not generate it with
  `ssh-keyscan` inside the deployment job
- `DEPLOY_ENV_FILE` — the complete environment-specific file based on
  `.env.production.example`
- `GHCR_USERNAME`, `GHCR_TOKEN` — a read-only package credential used by the VPS

The SSH account needs access to Docker and write access to `/opt/anochat`. Restrict
the SSH key to this deployment account and host.

## Normal deployment

1. Run `Deploy API` manually against `staging` for the candidate commit.
2. Confirm the health check and login/chat smoke flow in staging.
3. Merge the reviewed pull request into `main` and confirm all `CI` jobs pass.
4. Review and approve the pending `production` environment deployment.
5. Confirm the production health check, logs, and smoke flow.

The workflow does not deploy the Next.js frontend. Keep the frontend deployment
independent and set its production `NEXT_PUBLIC_API_URL` before building it.

## Rollback

1. Find the last known-good commit SHA in the package list or deployment history.
2. Run `Deploy API` manually, choose the affected environment, and enter that SHA
   in `image_tag`.
3. Approve the production deployment and confirm `/healthz` succeeds.
4. Verify authentication, queue join/leave, WebSocket connection, and message send.

Rollback changes the application image only. Database migrations must be
backward-compatible. Prefer additive migrations and forward fixes; never remove a
column in the same release that stops writing it.

## Failure runbook

- `/healthz` reports PostgreSQL unavailable: stop retrying deployments, verify the
  database provider and credentials, then restore from a tested backup if needed.
- `/healthz` reports Redis unavailable: matchmaking, sessions, rate limiting, and
  pub/sub are affected; restore Redis connectivity before accepting traffic.
- Migration fails: the API is not replaced. Inspect the one-shot `migrate` service
  logs, fix forward, publish a new SHA, and redeploy.
- API starts but the smoke flow fails: roll back to the last known-good SHA, retain
  logs, and reproduce in staging.

Test database backup restoration regularly; a backup that has never been restored
is not a verified recovery plan.

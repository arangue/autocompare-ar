# autocompare-ar

Scaffold for a used-car pricing and comparison app (Argentina).

## Layout

```text
cmd/server/     HTTP API
cmd/worker/     background jobs (stub)
internal/       domain packages (empty)
migrations/     SQL migrations
web/            Next.js frontend
```

## Quick start

```bash
cp .env.example .env
make db-up    # optional, for when DB is wired up
make dev      # API on :8080
make web      # frontend on :3000
```

GitHub remote: `git@github.com:arangue/autocompare-ar.git`

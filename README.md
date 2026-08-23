# autocompare-ar

Scaffold for a used-car pricing and comparison app (Argentina).

Full product and architecture design: [`docs/DESIGN.md`](docs/DESIGN.md).

## Architecture (clean / hexagonal)

Dependency rule: **inward only**. Outer layers depend on inner layers, never the reverse.

```text
cmd/server/main.go          composition root — wires everything

internal/delivery/http/     driving adapter  (HTTP handlers, router)
internal/application/       use cases        (orchestrate domain + ports)
internal/domain/            entities + ports (repository interfaces)
internal/infrastructure/    driven adapters  (postgres, ingestion, APIs)
```

```text
         HTTP request
              │
              ▼
    delivery/http/handler.go     ← parses request, calls use case
              │
              ▼
    application/list_brands.go   ← one use case per file (or small group)
              │
              ▼
    domain/ports.go              ← VehicleRepository interface
              ▲
              │ implements
    infrastructure/postgres/     ← SQL lives here only
```

| Layer | Add here |
|-------|----------|
| `domain/` | Entities (`Brand`, `Listing`, …) and port interfaces |
| `application/` | Use cases: `ListBrands`, `GetMarketSummary`, `AssessDeal`, … |
| `delivery/http/` | Handlers, router, middleware, JSON helpers |
| `infrastructure/postgres/` | Repository implementations + migrations wiring |
| `infrastructure/ingestion/` | CCA, ACARA, MercadoLibre importers |
| `migrations/` | SQL schema |
| `web/` | Next.js UI |

Handlers live in **`internal/delivery/http/`**. `main.go` creates repos → use cases → handler → router.

## Layout

```text
cmd/server/     HTTP API entrypoint
cmd/worker/     background jobs (stub)
internal/
  domain/       entities + ports
  application/  use cases
  delivery/http/ handlers + router
  infrastructure/postgres/
  infrastructure/ingestion/
migrations/
web/
```

## Quick start

```bash
cp .env.example .env
make db-up    # optional, for when DB is wired up
make dev      # API on :8080
make web      # frontend on :3000
```

GitHub remote: `git@github.com:arangue/autocompare-ar.git`

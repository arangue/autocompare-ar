# autocompare-ar

Tool for deciding whether a used car in Argentina is a good deal — not just another car search.

## Goal

Help buyers answer three questions:

1. **What is the real market price** of a given brand / model / year / trim?
2. **What does that trim actually include** (airbags, ABS, ESP, brakes, engine, etc.)?
3. **Is this listing cheap or expensive** relative to comparable cars?

The hero feature is deal assessment: given a published price, show how it sits against the market (e.g. “15% below median”). Reference guides (CCA, ACARA, DNRPA) and live listings (Mercado Libre, Autocosmos, etc.) feed that estimate — they are inputs, not the product.

MVP focus: a small catalog, market summary per trim/year, and a clear “is it cheap?” signal. No login, mobile app, microservices, or ML in v1.

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
make db-up    # Postgres on :5435
make dev      # API on :8080 (runs migrations)
make web      # frontend on :3000
```

API smoke check:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/brands
```

GitHub remote: `git@github.com:arangue/autocompare-ar.git`

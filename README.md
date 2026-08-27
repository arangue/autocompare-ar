# autocompare-ar

Tool for deciding whether a used car in Argentina is a good deal — not just another car search.

## Goal

Help buyers answer three questions:

1. **What is the real market price** of a given brand / model / year / trim?
2. **What does that trim actually include** (airbags, ABS, ESP, brakes, engine, etc.)?
3. **Is this listing cheap or expensive** relative to comparable cars?

The hero feature is deal assessment: given a published price, show how it sits against the market (e.g. “15% below median”). Reference guides (CCA, ACARA, DNRPA) and live listings (Mercado Libre, Autocosmos, etc.) feed that estimate — they are inputs, not the product.

MVP focus: a small catalog, market summary per trim/year, and a clear “is it cheap?” signal. No login, mobile app, microservices, or ML in v1.

Full product and architecture design: [DESIGN.md](DESIGN.md).

## API

Base URL (local): `http://localhost:8080`

### `GET /health`

Checks API + database connectivity.

**200**

```json
{"status":"ok"}
```

**503** — database unavailable

```json
{"code":"DB_UNAVAILABLE","message":"database unavailable"}
```



### `GET /api/v1/brands`

Lists brands from the catalog.

**200** — array of brands

```json
[
  {"ID":1,"Name":"Fiat","CreatedAt":"..."},
  {"ID":2,"Name":"Toyota","CreatedAt":"..."},
  {"ID":3,"Name":"Volkswagen","CreatedAt":"..."}
]
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to list brands"}
```

### `GET /api/v1/search/trims?q=&limit=`

Free-text search across brand, model, and trim names. Default `limit` is 20, max 50. Empty or whitespace `q` returns `[]`.

**200** — array of trims

```json
[
  {
    "trim_id": 1,
    "trim_name": "XEi 2.0 CVT",
    "brand_name": "Toyota",
    "model_name": "Corolla",
    "generation_name": "E210",
    "year_from": 2019,
    "year_to": 2026
  }
]
```

**400** — `limit` present but not a positive integer

```json
{"code":"INVALID_LIMIT","message":"limit must be a positive integer"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to search trims"}
```

These are the implemented endpoints only. Target MVP API surface: [DESIGN.md](DESIGN.md) §8.

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


| Layer                       | Add here                                                     |
| --------------------------- | ------------------------------------------------------------ |
| `domain/`                   | Entities (`Brand`, `Listing`, …) and port interfaces         |
| `application/`              | Use cases: `ListBrands`, `GetMarketSummary`, `AssessDeal`, … |
| `delivery/http/`            | Handlers, router, middleware, JSON helpers                   |
| `infrastructure/postgres/`  | Repository implementations + migrations wiring               |
| `infrastructure/ingestion/` | CCA, ACARA, MercadoLibre importers                           |
| `migrations/`               | SQL schema                                                   |
| `web/`                      | Next.js UI                                                   |


Handlers live in `internal/delivery/http/`. `main.go` creates repos → use cases → handler → router.

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

Postgres is required (the API pings the DB on `/health` and runs migrations on startup).

```bash
cp .env.example .env
make db-up    # Postgres on :5435
make dev      # API on :8080
make web      # frontend on :3000
```

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/brands
curl 'http://localhost:8080/api/v1/search/trims?q=corolla'
```

GitHub remote: `git@github.com:arangue/autocompare-ar.git`
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
  {"id":1,"name":"Fiat","created_at":"..."},
  {"id":2,"name":"Toyota","created_at":"..."},
  {"id":3,"name":"Volkswagen","created_at":"..."}
]
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to list brands"}
```

### `GET /api/v1/brands/{brand_id}/models`

Lists models for a brand. Unknown brand or no models returns `[]`.

**200** — array of models

```json
[
  {"id":1,"brand_id":2,"name":"Corolla","created_at":"..."}
]
```

**400** — `brand_id` not an integer

```json
{"code":"INVALID_ID","message":"brand_id must be an integer"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to list models by brand"}
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

### `GET /api/v1/trims/{trim_id}?year=`

Vehicle identity for one trim. `year` is optional and echoed as `requested_year` (must be a positive integer if present; not validated against listings).

**200**

```json
{
  "trim_id": 1,
  "trim_name": "XEi 2.0 CVT",
  "brand_id": 2,
  "brand_name": "Toyota",
  "model_id": 1,
  "model_name": "Corolla",
  "generation_id": 1,
  "generation_name": "E210",
  "year_from": 2019,
  "year_to": 2026,
  "requested_year": 2019
}
```

**400** — `trim_id` not an integer

```json
{"code":"INVALID_ID","message":"trim_id must be an integer"}
```

**400** — `year` present but not a positive integer

```json
{"code":"INVALID_YEAR","message":"year must be a positive integer"}
```

**404**

```json
{"code":"NOT_FOUND","message":"trim not found"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to get trim"}
```

### `GET /api/v1/trims/{trim_id}/listings?year=&limit=`

Active listings for a trim and year, ordered by price ascending. Default `limit` is 50, max 100.

**200** — array of listings (empty if none)

```json
[
  {
    "id": 1,
    "source": "seed",
    "external_id": "seed-corolla-xei-2019-01",
    "trim_id": 1,
    "year": 2019,
    "km": 120000,
    "price": 22500000,
    "currency": "ARS",
    "location": "Córdoba",
    "url": "https://example.com/listings/seed-corolla-xei-2019-01",
    "last_seen_at": "2026-08-29T00:00:00Z"
  }
]
```

**400** — `year` missing or invalid

```json
{"code":"INVALID_YEAR","message":"year is required"}
```

**400** — `limit` not a positive integer

```json
{"code":"INVALID_LIMIT","message":"limit must be a positive integer"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to list listings"}
```

### `GET /api/v1/trims/{trim_id}/market?year=`

Market price distribution from active listings for a trim and year.

**200** — with listings

```json
{
  "trim_id": 1,
  "year": 2019,
  "count": 10,
  "median": 26000000,
  "minimum": 22500000,
  "maximum": 29000000,
  "p25": 25200000,
  "p75": 27100000,
  "currency": "ARS"
}
```

**200** — no listings (`count` 0, percentiles `null`)

```json
{
  "trim_id": 1,
  "year": 2019,
  "count": 0,
  "median": null,
  "minimum": null,
  "maximum": null,
  "p25": null,
  "p75": null,
  "currency": "ARS"
}
```

**400** — `year` missing or invalid

```json
{"code":"INVALID_YEAR","message":"year is required"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to get market summary"}
```

### `GET /api/v1/trims/{trim_id}/references?year=`

Guide and fiscal reference prices (CCA, ACARA, DNRPA) for a trim and year. DNRPA is labeled `kind: "fiscal"`; guides use `kind: "guide"`. These are estimative inputs, not market prices.

**200** — array of references (empty if none)

```json
[
  {
    "id": 1,
    "trim_id": 1,
    "year": 2019,
    "source": "CCA",
    "kind": "guide",
    "price": 25500000,
    "currency": "ARS",
    "observed_at": "2026-08-29T00:00:00Z"
  },
  {
    "id": 3,
    "trim_id": 1,
    "year": 2019,
    "source": "DNRPA",
    "kind": "fiscal",
    "price": 24100000,
    "currency": "ARS",
    "observed_at": "2026-08-29T00:00:00Z"
  }
]
```

**400** — `year` missing or invalid

```json
{"code":"INVALID_YEAR","message":"year is required"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to list references"}
```

### `GET /api/v1/trims/{trim_id}/deal?year=&price=`

Compare an asking price to the market median and percentile band for a trim and year. Hero endpoint for “¿está barato?”.

**200** — enough listings

```json
{
  "trim_id": 1,
  "year": 2019,
  "price": 21500000,
  "currency": "ARS",
  "status": "ok",
  "band": "below",
  "delta_ars": 4500000,
  "delta_pct": 17.3,
  "market": {
    "count": 10,
    "median": 26000000,
    "p25": 25200000,
    "p75": 27100000
  },
  "disclaimer": "Estimación a partir de publicaciones comparables; el precio publicado no es precio de venta."
}
```

`delta_ars` is `median - price` (positive means cheaper than median). `delta_pct` is `(median - price) / median * 100`, one decimal. `band` is `below` if price &lt; p25, `near` if between p25 and p75, `above` if &gt; p75.

**200** — insufficient data (`status` and `band` are `insufficient_data`, deltas `null`)

```json
{
  "trim_id": 1,
  "year": 2019,
  "price": 21500000,
  "currency": "ARS",
  "status": "insufficient_data",
  "band": "insufficient_data",
  "delta_ars": null,
  "delta_pct": null,
  "market": {
    "count": 0,
    "median": null,
    "p25": null,
    "p75": null
  },
  "disclaimer": "No hay publicaciones suficientes para estimar el mercado de esta versión/año."
}
```

**400** — `year` or `price` missing or invalid

```json
{"code":"INVALID_PRICE","message":"price is required"}
```

**500**

```json
{"code":"INTERNAL_ERROR","message":"failed to assess deal"}
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
curl http://localhost:8080/api/v1/brands/2/models
curl 'http://localhost:8080/api/v1/search/trims?q=corolla'
curl 'http://localhost:8080/api/v1/trims/1?year=2019'
curl 'http://localhost:8080/api/v1/trims/1/listings?year=2019'
curl 'http://localhost:8080/api/v1/trims/1/market?year=2019'
curl 'http://localhost:8080/api/v1/trims/1/references?year=2019'
curl 'http://localhost:8080/api/v1/trims/1/deal?year=2019&price=21500000'
```

GitHub remote: `git@github.com:arangue/autocompare-ar.git`
# AutoCompare AR — Design Document

Exported from the product/architecture design conversations for `autocompare-ar`.

This document captures the **core idea**, **data sources**, **MVP scope**, **tech recommendations**, **data model**, **product features**, **risks**, **build order**, and the **clean / hexagonal architecture** used in the repo.

---

## 1. Core idea

**AutoCompare AR** is not “another used-car search.”

It is a **purchase decision tool** for the Argentine used-car market.

Buyers (and later dealers) should answer three questions:

1. **What is the real market price** of a given brand / model / year / trim?
2. **What does that version actually include?**  
   (airbags, ABS, ESP, disc brakes, cruise control, ISOFIX, engine, transmission, consumption, etc.)
3. **Is this listing cheap or expensive** relative to comparable cars?

### Hero feature

> **“Encontrá si estás pagando de más.”**  
> (“Find out if you’re overpaying.”)

Example:

```text
Toyota Corolla XEi 2.0 CVT 2019

PRECIO DE MERCADO          $25.300.000
Publicación encontrada     $21.500.000

🟢 $3.800.000 debajo del mercado
🟢 15% más barato
```

That is more valuable than showing only `$21.500.000`.

### Product framing

| Not this | This |
|----------|------|
| Another catalog / scraper UI | Decision tool |
| “Search all cars in Argentina” | Market price + specs + deal signal |
| One “official” price | Several prices + a market distribution |
| Listing-first model | Canonical vehicle identity first |

Working product name from design: **AutoCompare Argentina**.

---

## 2. Price and data sources

Do **not** try to obtain “the” price. Obtain **several** prices and present them with clear meaning.

### Reference / guide sources

| Source | Role | How to show it |
|--------|------|----------------|
| **CCA** (Cámara del Comercio Automotor) | Monthly guide of used and 0 km prices by brand/model/year/version | Reference price |
| **ACARA** | Official used-vehicle valuation guide from observed commercial behavior | Reference price |
| **DNRPA** | Values used for registration / transfer / fees | **Fiscal / registral valuation** — not market price |

### Market / listing sources

| Source | Role |
|--------|------|
| **Mercado Libre** | Real published listing prices; official vehicle APIs exist (check allowed use cases) |
| **Autocosmos** | Listings + technical sheets / equipment comparator |
| Other classifieds | Additional market observations |

### Market math (from listings)

For a trim + year (example: Toyota Corolla XEi 2.0 CVT 2019), collect listing prices and compute:

- count
- median
- min / max
- p25 / p75

Example:

```text
median = $26.0M
p25    = $25.2M
p75    = $27.1M

Listing at $22.5M → 🟢 ~13.5% below market
```

### Important caveats

- Published price ≠ sold price.
- Guide values are estimative and vary with condition, location, km, etc.
- Always label sources clearly (CCA vs ACARA vs DNRPA vs market median).

### Conceptual price table per vehicle

| Fuente | Precio |
|--------|--------|
| CCA | $XX |
| ACARA | $XX |
| DNRPA | $XX (fiscal) |
| Mercado publicado (median) | $XX |
| Autocosmos / other | $XX |
| Tu estimación | $XX |

---

## 3. Hard problem: versions / identity

The difficult problem is **not** the scraper. It is **vehicle canonicalization**.

`Toyota Corolla 2019` is not enough. You may have:

- Corolla XLi
- Corolla XEi
- Corolla SEG
- Corolla Hybrid
- …

Worse: the same trim appears differently across sources:

- Toyota Corolla XEI 2.0 CVT
- Toyota Corolla XEi Pack
- Toyota Corolla 2.0 XEI CVT
- Toyota Corolla XEI 2.0

You need a **vehicle identity model**:

```text
brand → model → generation → trim → specs / features
```

Example:

```text
Toyota
 └── Corolla
      └── E210
           └── XEi
                ├── ABS = true
                ├── ESP = true
                ├── airbags = 7
                ├── front_disc = true
                ├── rear_disc = true
                └── cruise_control = true
```

That enables queries like:

> “Show cars 2018–2021 with at least 6 airbags and ESP under $25M.”

### Critical separation

Do **not** mix:

- **Vehicle** = conceptual trim/version (Toyota Corolla XEi 2019)
- **Listing** = a specific ad (that trim, 120.000 km, $22.5M, Córdoba, Mercado Libre)

```text
Vehicle
   │
   ├── Listing
   ├── Listing
   ├── Listing
   └── Listing
```

Listings feed market stats. Vehicles own specs and identity.

---

## 4. MVP scope

### In scope (v1)

**A. Search**
- Brand / model / year / version (trim)

**B. Vehicle sheet**
- Motor, transmisión, combustible, potencia, consumo
- Seguridad: airbags, ABS, ESP, ISOFIX
- Frenos: delanteros / traseros
- Equipamiento: cruise, cámara, sensores, climatizador, …

**C. Reference prices**
- CCA / ACARA / market (and DNRPA labeled as fiscal)

**D. Market summary**
- Count, median, min, max, p25, p75

**E. Deal assessment (“¿Está barato?”)** — likely the differentiator

**F. Comparator**
- Side-by-side trim comparison (price + equipment)

### Out of scope (v1)

- Login / accounts
- Mobile app
- Microservices
- ML / AI recommendations
- Scraping dozens of sites
- Kafka / Redis / K8s / API Gateway / Lambdas “because architecture”

### Catalog size for first version

Start small, expand later:

- ~**100–300** models/versions
- ~**10 brands** first (e.g. Toyota, Volkswagen, Fiat, Chevrolet, Ford, Renault, Peugeot, Citroën, Honda, Nissan)
- Years roughly **2015 → 2026**

Then grow: 10 → 30 → 50 → all.

### Current seed in the repo

Already seeded for development:

| Brand | Model | Generation | Trim(s) |
|-------|-------|------------|---------|
| Toyota | Corolla | E210 (2019–2026) | XEi 2.0 CVT, XLi 1.8 CVT |
| Volkswagen | Golf | Mk7 (2014–2020) | Comfortline |
| Fiat | Cronos | 1st gen (2018–2026) | Drive 1.3 |

---

## 5. Product UX sketches (from design)

### Search → detail

```text
┌───────────────────────────────────────────────┐
│ Buscar auto                                   │
│ Toyota Corolla XEi 2019                       │
└───────────────────────────────────────────────┘
                      ↓
┌───────────────────────────────────────────────┐
│ Toyota Corolla XEi 2.0 CVT 2019               │
│ Precio mercado        $26.0M                  │
│ CCA                    $25.5M                 │
│ Publicaciones          $25M - $29M            │
│ 🟢 Precio por debajo del mercado              │
└───────────────────────────────────────────────┘

 Seguridad / Frenos / Motor  (feature sections)

 PUBLICACIONES ACTUALES
 $22.5M  120.000 km  Córdoba
 $24.8M   98.000 km  Rosario
 ...
```

### Opportunity / margin view (later feature)

```text
OPORTUNIDAD
Precio publicado       $21.5M
Precio mercado         $25.3M
Diferencia             +$3.8M (+17.7%)

Warning: estimate from comparable publications;
not a guarantee you can sell at market median.
```

### Monetization (later)

**Free**
- Search, sheet, reference prices, basic comparator

**Premium**
- Price history, alerts, “cheap car found”, market evolution, advanced comparison

**Leads**
- Dealer leads from opportunity lists

---

## 6. Tech recommendations

### Architecture style

**Modular monolith** — one deployable API, clear internal packages.

Not microservices for v1. Domain is not stable enough to justify that complexity.

### Recommended stack (design)

| Layer | Choice |
|-------|--------|
| Backend | **Go** (`net/http`, `pgx`, migrations) |
| Database | **PostgreSQL** (not MongoDB) |
| Frontend (fast MVP path) | Go + HTMX + Tailwind |
| Frontend (chosen for this repo) | **Go API + Next.js (React) + Tailwind** |
| Background work | Go worker process |
| Local DB | Docker Compose Postgres |

Repo decision: **Go API + React/Next.js + PostgreSQL**, matching `github.com/arangue/autocompare-ar`.

### Process layout

```text
Browser (web/)
      │
      ▼
Go HTTP API (cmd/server)
      │
      ├── vehicle / pricing / listing / search use cases
      └── PostgreSQL
              ▲
              │
Go Worker (cmd/worker)
      └── ingestion (CCA, ACARA, classifieds, specs)
```

### What not to add early

- Kafka / RabbitMQ
- Kubernetes / EKS
- Redis (until measured need)
- API Gateway
- Lambda / 10 microservices

### Suggested repo structure (implemented)

```text
autocompare-ar/
├── cmd/server/                 # HTTP API entrypoint (composition root)
├── cmd/worker/                 # background jobs (stub)
├── internal/
│   ├── domain/                 # entities + ports
│   ├── application/            # use cases
│   ├── delivery/http/          # handlers, router, JSON helpers
│   └── infrastructure/
│       ├── postgres/           # repository adapters
│       └── ingestion/          # future importers
├── migrations/                 # SQL schema + seed
├── web/                        # Next.js frontend
├── docker-compose.yml
├── Makefile
└── docs/DESIGN.md              # this document
```

### Ingestion pipelines (future)

Separate **spec ingestion** from **price/listing ingestion**.

```text
Specs: Autocosmos / manufacturers / other → Normalizer → PostgreSQL
Prices: CCA / ACARA / classifieds → Price ingest → Normalization → PostgreSQL
```

Store each listing observation over time (`first_seen_at`, `last_seen_at`) so you can later show:

- “Listed at $25M 40 days ago, now $23M”
- “Corolla XEi 2019 median fell 4.2% in 60 days”

### Similar-trim weighting (later)

When exact trim/year sample is thin:

| Match | Weight |
|-------|--------|
| Exact trim/year | 1.0 |
| Same trim ±1 year | 0.8 |
| Same model/year | 0.7 |
| Same segment | 0.4 |

Do this after core market summary works.

---

## 7. Data model

### Conceptual hierarchy

```text
brand
  id, name

model
  id, brand_id, name

generation
  id, model_id, name, year_from, year_to

trim
  id, generation_id, name, year_from, year_to

vehicle_spec
  trim_id, engine, displacement, horsepower, transmission,
  fuel, doors, seats, consumption_city, consumption_highway

features / vehicle_features
  feature catalog + trim ↔ feature values

vehicle_listing
  id, source, external_id, trim_id, year, km, price, currency,
  location, url, first_seen_at, last_seen_at, active

price_references
  id, trim_id, year, source, price, currency, observed_at
```

### Implemented today (`migrations/001_create_catalog`)

Catalog only:

- `brands`
- `models`
- `generations`
- `trims`

### Planned next tables

- `features`, `vehicle_specs`, `vehicle_features`
- `vehicle_listings`
- `price_references`

Indexes (when listings exist):

- `(trim_id, year)` on active listings
- `price` on active listings
- `(trim_id, year)` on price references

### Domain entities already sketched in code

- `Brand`, `Model`, `Trim`
- `Listing`, `MarketSummary`
- `ReferencePrice`
- Ports: `VehicleRepository`, `ListingRepository`, `PricingRepository`

---

## 8. API surface (target MVP)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Health (+ DB ping) |
| GET | `/api/v1/brands` | List brands |
| GET | `/api/v1/brands/{brandID}/models` | List models |
| GET | `/api/v1/search/trims?q=` | Search trims |
| GET | `/api/v1/trims/{trimID}?year=` | Vehicle detail |
| GET | `/api/v1/trims/{trimID}/listings?year=` | Active listings |
| GET | `/api/v1/trims/{trimID}/market?year=` | Market summary |
| GET | `/api/v1/trims/{trimID}/references?year=` | Guide prices |
| GET | `/api/v1/trims/{trimID}/deal?year=&price=` | Deal assessment |

### Implemented today

- `GET /health`
- `GET /api/v1/brands` (reads Postgres; seeded brands)

---

## 9. Legal / data risk

Largest project risk is **data licensing / ToS**, not Go.

Preference order:

1. Official API — preferred  
2. Public downloadable / licensed data — good  
3. Scraping — check ToS / robots / licensing  
4. Copying entire third-party databases — avoid  

Notes from design research:

- CCA / ACARA content often has reproduction restrictions.
- ACARA in particular indicates prohibition of total/partial reproduction of its information.

Do not build the whole business on republishing a third-party guide wholesale. Prefer:

- licensed feeds
- your own market observations from allowed sources
- clear attribution / non-verbatim aggregation where legally appropriate

(This is product guidance, not legal advice — review before commercial launch.)

---

## 10. Recommended build order

| Phase | Deliverable |
|-------|-------------|
| 1 | DB + migrations + catalog seed ✅ |
| 2 | Real `ListBrands` (+ wiring) ✅ |
| 3 | Models / search / trim detail endpoints |
| 4 | Listings table + market summary |
| 5 | Deal assessment endpoint |
| 6 | Next.js search + vehicle + deal badge UI |
| 7 | Worker + ingestion (one source first) |
| 8 | Comparator, history, alerts |

Rule: ship a thin vertical slice each time (schema → port → SQL → use case → handler → route → `main` wiring).

---

## 11. Clean / hexagonal architecture

Clean Architecture and Hexagonal Architecture (ports & adapters) are the same idea here:

> **Domain and use cases sit in the center. Adapters plug into the edges. Dependencies point inward.**

### Layers

```text
                    ┌─────────────────────────┐
   Browser (web/)   │   delivery/http         │  ← driving adapter (HTTP in)
                    └───────────┬─────────────┘
                                │ calls
                    ┌───────────▼─────────────┐
                    │   application           │  ← use cases
                    └───────────┬─────────────┘
                                │ uses
                    ┌───────────▼─────────────┐
                    │   domain                │  ← entities + ports (interfaces)
                    └───────────▲─────────────┘
                                │ implements
                    ┌───────────┴─────────────┐
                    │   infrastructure        │  ← driven adapters (Postgres, scrapers)
                    └─────────────────────────┘
```

### What each layer owns

| Layer | Path | Responsibility |
|-------|------|----------------|
| Composition root | `cmd/server/main.go` | Wire config, DB, repos, use cases, handlers, router. No business logic. |
| Driving adapter | `internal/delivery/http/` | Parse HTTP, call use cases, write JSON/status. |
| Application | `internal/application/` | One use case per action (`ListBrands`, later `AssessDeal`, …). |
| Domain | `internal/domain/` | Entities + **ports** (repository interfaces). No SQL, no HTTP. |
| Driven adapters | `internal/infrastructure/postgres/` | Implement ports with SQL. |
| Ingestion adapters | `internal/infrastructure/ingestion/` | Import CCA/ACARA/classifieds into DB via repos. |
| Worker entrypoint | `cmd/worker/` | Run ingestion / scheduled jobs. |
| UI | `web/` | Next.js client of the API. |

### Ports vs adapters

- **Port** = interface the core needs (`VehicleRepository` in `domain/ports.go`).
- **Adapter** = concrete implementation (`postgres.VehicleRepository`).

Application code depends on the **port**, never on Postgres directly.

### Request flow (example: list brands)

```text
GET /api/v1/brands
        │
        ▼
delivery/http/router.go     route → Handler.ListBrands
        │
        ▼
delivery/http/handler.go    call listBrands.Execute(ctx)
        │
        ▼
application/list_brands.go  call repo.ListBrands(ctx)
        │
        ▼
domain/ports.go             VehicleRepository interface
        ▲
        │ implements
infrastructure/postgres/vehicle_repository.go
        │
        ▼
PostgreSQL  SELECT id, name, created_at FROM brands
        │
        ▼
JSON response back up the chain
```

### Why this shape

- Swap Postgres for another store without rewriting use cases (as long as the port holds).
- Test use cases with fake repositories.
- Keep HTTP details out of domain.
- Keep SQL out of handlers.
- Match the style already used in your `wallet-service` projects.

### Adding a new endpoint (recipe)

Example: `GET /api/v1/trims/{id}/market`

1. Extend / confirm port in `domain/ports.go`
2. Implement SQL in `infrastructure/postgres/`
3. Add use case in `application/`
4. Add handler method in `delivery/http/handler.go`
5. Register route in `delivery/http/router.go`
6. Wire in `cmd/server/main.go`

### What not to do

- Business rules in handlers
- SQL in `main.go` or handlers
- Domain importing `net/http` or `pgx`
- Mixing vehicle identity rows with listing rows
- Putting scrapers inside `domain/` (they belong in `infrastructure/ingestion/`)

### Mental model

- **Handler** = waiter  
- **Use case** = kitchen ticket  
- **Domain** = menu / recipes  
- **Infrastructure** = pantry / database  

The waiter never goes into the pantry. The ticket goes to the kitchen; the kitchen uses the pantry through a defined interface.

---

## 12. Status snapshot (as of this document)

### Done

- Repo scaffold (Go API + Next.js + Docker Postgres)
- Clean/hex folder layout
- Catalog migrations + tiny seed
- Postgres wiring + migrations on startup
- `GET /health`, `GET /api/v1/brands`

### Not done yet

- Models / trim search / vehicle detail APIs
- Listings + market summary + deal assessment
- Specs / features tables
- Ingestion workers
- Real frontend beyond default Next.js template
- Comparator UI
- Auth / monetization

---

## 13. Quick start (local)

```bash
cd ~/go/src/github.com/arangue/autocompare-ar   # or your local clone
cp .env.example .env
make db-up
make dev      # API :8080
make web      # frontend :3000
```

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/brands
```

---

## 14. Sources / provenance

Product design synthesized from:

- Shared ChatGPT design thread: **“Diseñar MVP cotizaciones autos”**  
  (`https://chatgpt.com/s/t_6a838f4190e08191bc163bca724b3618`)
- Follow-up decisions in this Cursor conversation:
  - stack = Go API + React/Next.js + PostgreSQL
  - architecture = clean / hexagonal (wallet-service style)
  - handlers in `internal/delivery/http/`
  - composition root in `cmd/server/main.go`
  - start with catalog migration + seed + ListBrands

External references discussed in the design chat (for research / validation, not copied wholesale):

- CCA price guides
- ACARA official price guide
- DNRPA valuation tables (fiscal/registral)
- Autocosmos listings / tech sheets
- Mercado Libre Developers vehicle docs

---

*End of design document.*

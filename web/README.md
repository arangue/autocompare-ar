# AutoCompare web

Next.js frontend for AutoCompare AR.

## Setup

```bash
cp .env.example .env.local
npm install
```

`NEXT_PUBLIC_API_URL` defaults to `http://localhost:8080`. The dev server proxies `/api/*` to that URL so browser requests avoid CORS.

## Run

Start the Go API first (`make api` from the repo root), then:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

Search a trim on the home page, then open `/trims/{id}?year=` for market data, references, specs, listings, and the deal calculator.

## API client

Typed helpers live in `src/lib/api.ts`:

- `searchTrims`
- `getTrim`
- `getMarket`
- `getDeal`
- `getListings`
- `getReferences`
- `getCompare`

Types mirror the Go API JSON in `src/lib/types.ts`.

Compare page: `/compare?trim_ids=1,2&year=2019` (add trims via **Comparar** on the vehicle sheet).

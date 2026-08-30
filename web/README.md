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

## API client

Typed helpers live in `src/lib/api.ts`:

- `searchTrims`
- `getTrim`
- `getMarket`
- `getDeal`
- `getListings`
- `getReferences`

Types mirror the Go API JSON in `src/lib/types.ts`.

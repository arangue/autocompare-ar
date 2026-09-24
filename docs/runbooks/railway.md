# Railway

## Servicios (3, siempre prendidos)

- Postgres (plugin)
- API: Dockerfile de la raíz del repo (Go)
- Web: Root Directory = `web`, Dockerfile `web/Dockerfile`
  - Dónde: servicio web → Settings → Source → Add Root Directory

## Worker (cron u one-off, no es HTTP)

No uses el Dockerfile de la raíz ni Generate Domain: Railway le pega healthcheck a `/` y mata el proceso (`exited`) porque el worker no escucha HTTP.

- Dockerfile path: `Dockerfile.worker` (Settings → Build)
- Env: `DATABASE_URL=${{NOMBRE_CAJITA_POSTGRES.DATABASE_URL}}` y `INGEST_FILE=/testdata/listings.json`
- No pongas `EXPOSE` ni healthcheck HTTP en este servicio

## Variables

- API: solo `DATABASE_URL=${{NOMBRE_CAJITA_POSTGRES.DATABASE_URL}}`
  - El nombre `NOMBRE_CAJITA_POSTGRES` tiene que ser el de la cajita (a menudo `Postgres`)
- Web (BUILD): `NEXT_PUBLIC_API_URL=https://<domain-de-la-API>`
- No pongas `INGEST_FILE` ni `NEXT_PUBLIC` en la API
- No uses localhost en prod

## Comprobar

- API: `https://<api>/health` → `{"status":"ok"}`
- Web: `https://<web>/` es la portada AutoCompare
- Si `/health` ok pero `/` es 404 JSON: estás en el domain de la API

## Si exited

- Logs: `DATABASE_URL is not set` o unix `/tmp/.s.PGSQL.5432` = URL rota / nombre de cajita mal

# Cursor Prompts — Ingestion Live (autocompare-ar)

Usá estos prompts **uno por uno** en Cursor.  
No pases al siguiente hasta que el anterior esté mergeado / funcionando.

Convenciones:
- Repo: `arangue/autocompare-ar`
- Arquitectura hexagonal ya existente: no romper capas.
- Reusar: `ListingRepository.Upsert`, `AliasRepository.Resolve`, `normalize.Title`, worker actual.
- Tests obligatorios cuando haya lógica nueva.
- No inventar fuentes ni endpoints que no existan.

---

## PROMPT 0.2 — Expandir trim_aliases (bloqueante)

```
Contexto: proyecto Go autocompare-ar (clean/hexagonal). Worker de ingestion ya existe en cmd/worker y resuelve raw_title vía AliasRepository.Resolve + normalize.Title.

Tarea: expandir trim_aliases para los trims seedados, sin tocar todavía Mercado Libre.

1. Revisá migrations/006_create_trim_aliases.up.sql y los trims seedados en migrations (Corolla XEi/XLi, Golf Comfortline, Cronos Drive).
2. Creá migrations/009_seed_trim_aliases_expand.up.sql y .down.sql.
3. Insertá aliases con raw_name ya normalizado (UPPER, sin tildes, espacios colapsados), source = 'file' o 'mercadolibre' según convenga, apuntando a los trim_id correctos.
4. Cubrí variantes típicas de títulos de clasificados:
   - Toyota Corolla XEi 2.0 CVT / XEi 2.0 / XEi / XEi Pack
   - Toyota Corolla XLi 1.8 CVT / XLi
   - Volkswagen Golf Comfortline / Golf Comfortline
   - Fiat Cronos Drive 1.3 / Cronos Drive
5. No cambies el schema de trim_aliases.
6. Verificá que normalize.Title del repo produce el mismo string que guardás en raw_name.
7. Corré: go run ./cmd/worker -dry-run testdata/listings.titles.json
8. Criterio de done: ≥80% de los títulos de prueba de esos trims se resuelven. Si hace falta, ampliá testdata/listings.titles.json con más ejemplos realistas.

No implementes cliente de Mercado Libre en este paso.
```

---

## PROMPT 1.1 — Investigar API Mercado Libre (manual / notas)

```
Contexto: autocompare-ar. Vamos a ingerir listings de Mercado Libre Argentina vía API oficial.

Tarea (investigación, sin código de producción todavía):

1. Documentá en docs/ml-api-notes.md (crealo):
   - Cómo registrar app en developers.mercadolibre.com.ar
   - Endpoint de búsqueda de vehículos para MLA
   - category id recomendado para autos (verificar si MLA1744 sigue vigente)
   - Campos del JSON de search/items útiles para nosotros: id, title, price, currency_id, permalink, attributes (año, km, etc.), ubicación
   - Si la búsqueda pública necesita access token o no
   - Rate limits conocidos
   - Ejemplo real de curl que funcione hoy

2. No escribas todavía el package mercadolibre. Solo notas + ejemplos de respuesta (redactá tokens).

3. Actualizá .env.example con placeholders:
   ML_ACCESS_TOKEN=
   ML_SITE_ID=MLA

Criterio de done: docs/ml-api-notes.md con curl de ejemplo y mapeo de campos claros.
```

---

## PROMPT 1.2 — Package mercadolibre (client + mapper)

```
Contexto: autocompare-ar. Ya existen:
- domain.Listing
- normalize.Title
- AliasRepository / ListingRepository
- Worker que hace Resolve + Upsert

Tarea: crear el adapter de Mercado Libre SIN cablearlo al worker todavía.

Creá:

internal/infrastructure/ingestion/mercadolibre/
  client.go      // HTTP client, timeout, rate limit simple
  search.go      // Search(ctx, query, limit, offset)
  mapper.go      // RawItem → domain.Listing
  client_test.go
  mapper_test.go

Reglas:
1. source fijo = "mercadolibre"
2. external_id = id del ítem ML (ej. MLA123...)
3. raw_title = title del ítem
4. NO setear trim_id en el mapper (lo resuelve el worker con aliases)
5. price, currency, url (permalink), year/km desde attributes si existen, location si existe
6. Si falta year o price válidos, el mapper puede devolver ok=false o el caller skipea (documentá la convención)
7. Site MLA por defecto; leer ML_SITE_ID y ML_ACCESS_TOKEN del env si hace falta
8. Rate limit conservador (ej. mínimo 1s entre requests o configurable)
9. Logs con log/slog
10. Tests unitarios del mapper con JSON de ejemplo realista (fixtures en testdata o inline)
11. No dependas de postgres ni del worker en este package

Criterio de done: tests del mapper en verde; client compilable; se puede llamar Search desde un main temporal o test de integración opcional con httptest.
```

---

Antes del PROMPT 1.3, cerrá E5-502 y E5-503. El flag `MARKET_EXCLUDE_SEED` queda en `false`: el demo sigue contando el seed. El primer upsert real de Corolla XEi 2019 mezcla seed y live si el flag no existe en el código.

## PROMPT 1.3 — Cablear ML en el worker

```
Contexto: autocompare-ar.
- cmd/worker ya parsea archivos JSON/CSV, resuelve aliases y hace Upsert/Expire.
- Existe internal/infrastructure/ingestion/mercadolibre (client + mapper).

Tarea: extender el worker para soportar fuente Mercado Libre sin romper el flujo por archivo.

1. Agregá flag -source (default vacío o "file"):
   - sin -source o archivo como arg → comportamiento actual (ParseFile)
   - -source mercadolibre → usa el client ML

2. Para mercadolibre:
   - Lista inicial de queries (const o config):
     toyota corolla xei
     toyota corolla xli
     volkswagen golf comfortline
     fiat cronos drive
   - Por cada query: Search paginado (ej. hasta 50–100 items)
   - Mapear a []domain.Listing
   - Reutilizar el loop actual: si TrimID==0 → AliasRepository.Resolve(source, normalize.Title(raw_title)) → Upsert
   - Respetar -dry-run y -expire-days

3. Coexistencia:
   go run ./cmd/worker testdata/listings.json
   go run ./cmd/worker -source mercadolibre -dry-run
   go run ./cmd/worker -source mercadolibre -expire-days 14

4. Errores de red/API: loggear y salir con código != 0 si falla de forma dura; skip de ítems individuales si el mapper no puede mapear.

5. Actualizá README.md sección Worker ingest con los nuevos flags.

6. No implementes CCA/ACARA en este paso.

Criterio de done:
- dry-run contra ML real (o mock) muestra matches/skips
- upsert local aumenta vehicle_listings
- GET /api/v1/trims/{id}/market?year=... incluye las filas nuevas
- MARKET_EXCLUDE_SEED sigue en false. No lo pongas en true en este paso (eso es E5-504, con ≥5 listings que no son seed)
```

---

## PROMPT 1.4 — Hardening ML (calidad + secrets)

```
Contexto: worker ya puede -source mercadolibre.

Tarea: robustecer la ingestion de ML.

1. Validaciones antes de Upsert:
   - price > 0
   - year entre 1990 y año actual+1
   - currency default ARS si viene vacío
2. Skip + log warn si no pasa validación.
3. Contadores claros al final (igual que ahora): matched, inserted, updated, skipped, expired.
4. Asegurar que .env.example documenta ML_* y que ningún token queda hardcodeado.
5. Rate limit y timeout documentados en comentario del client.
6. Si ML devuelve 0 resultados para todas las queries, log de error explícito.

Criterio de done: corrida local estable; listings basura no entran; README actualizado.
```

---

## PROMPT 2.1 — Observabilidad de unmatched titles

```
Contexto: autocompare-ar. La ingestion de ML skipea títulos sin alias.

Tarea: hacer visibles los no-matches para mejorar aliases.

1. Cuando Resolve devuelve ok=false, además del slog.Warn:
   - escribir a un archivo de log (ej. logs/unmatched-titles.jsonl) o
   - insertar en tabla simple unmatched_titles (source, raw_title, seen_at, count) si preferís DB
2. Elegí la opción más simple que no rompa la arquitectura (archivo es suficiente para MVP).
3. Documentá en README cómo revisar unmatched y agregar aliases (migración o SQL manual).

Criterio de done: después de una corrida ML, tengo una lista accionable de títulos a aliasar.
```

---

## PROMPT 2.2 — Más aliases desde unmatched reales

```
Contexto: ya corrimos ML y tenemos unmatched titles.

Tarea:
1. Tomá los unmatched más frecuentes de los trims seedados.
2. Creá migration 010 (o siguiente número) con nuevos trim_aliases.
3. Solo aliases de alta confianza (misma versión clara).
4. Re-corrê dry-run / ingest y medí mejora de % matched.

No expandas el catálogo de trims todavía salvo que un unmatched demuestre un trim seedado mal nombrado.
```

---

## PROMPT 2.3 — Expandir catálogo (solo si hace falta)

```
Contexto: autocompare-ar. Catálogo seed pequeño (Corolla, Golf, Cronos).

Tarea (solo si el matching de ML ya está >60-70% en seed):
1. Agregar 1–2 modelos populares más vía migrations (ej. Toyota Hilux una versión, VW Amarok una versión, o Fiat Cronos otra versión).
2. Incluir: generation, trim, specs mínimas si aplica, price_references seed opcionales, y aliases ML típicos.
3. Seguir el estilo de migrations/007 y 008.

No agregues 50 modelos. Máximo un slice vertical utilizable.
```

---

## PROMPT 3.1 — Ingestion CCA (referencias, no listings)

```
Contexto: autocompare-ar.
- price_references ya existe (trim_id, year, source, kind, price, ...)
- UNIQUE (trim_id, year, source)
- PricingRepository.ListReferences ya se usa en la API
- Listings de ML ya funcionan

Tarea: ingestion de guía CCA como referencias, SEPARADA de listings.

1. Package internal/infrastructure/ingestion/cca/
   - parser de PDF o CSV/JSON intermedio (lo más simple y estable)
   - mapper → estructura para upsert en price_references (source="CCA", kind="guide")
2. Matching a trim_id: por aliases o tabla de mapeo explícita marca/modelo/versión → trim_id (documentá limitaciones).
3. Worker: flag -source cca o subcomando; NO mezclar con -source mercadolibre.
4. Upsert idempotente mensual.
5. Legal: no guardar texto completo de la guía; solo precios + metadatos mínimos.

Criterio de done: corrida carga/actualiza price_references; /api/v1/trims/{id}/references?year= muestra CCA real además del seed.
```

---

## PROMPT 3.2 — Ingestion ACARA

```
Contexto: igual que CCA. price_references + worker de referencias.

Tarea: adapter ACARA análogo a CCA.
- source="ACARA", kind="guide"
- Preferir PDF bulk si existe; evitar scrapear el wizard multi-paso si es frágil
- Mismo criterio legal: no republicar la guía completa
- Flag -source acara o mismo worker de referencias

Criterio de done: referencias ACARA visibles en la API sin romper CCA ni listings.
```

---

## PROMPT 4.1 — Scheduling y operación

```
Contexto: ML + (opcional) CCA/ACARA ya corren a mano.

Tarea:
1. Documentá en docs/runbooks/ingestion.md:
   - comandos exactos para ML diario, expire, CCA/ACARA mensual
   - variables de entorno necesarias
   - qué logs mirar
2. Si el deploy es Railway (ver docs/runbooks/railway.md), proponé cron jobs concretos o script Makefile (make ingest-ml, make ingest-expire).
3. No implementes Kubernetes ni colas nuevas.

Criterio de done: un humano puede operar ingestion leyendo solo el runbook.
```

---

## PROMPT 4.2 — Monitoreo mínimo

```
Contexto: worker loguea matched/inserted/updated/skipped/expired.

Tarea:
1. Asegurar JSON logs estables (campos fijos) para poder grepear.
2. Al final de cada corrida, log nivel Error si:
   - todas las queries ML devolvieron 0 items
   - skipped > 50% de matched (umbral configurable)
3. Query SQL de ejemplo en el runbook: listings activos por trim/year.

Sin Prometheus/Grafana todavía.
```

---

## Cómo usar esto en Cursor

1. Abrí el repo en Cursor.
2. Pegá **un solo PROMPT** en el chat (Agent/Composer).
3. Pedile que implemente solo ese alcance.
4. Corré tests / worker / API localmente.
5. Commit.
6. Recién ahí el siguiente prompt.

Orden obligatorio al inicio. Cada paso es una tarjeta del tablero:

1. E10-002 · PROMPT 0.2 (hecho)
2. E10-011 · PROMPT 1.1 (hecho)
3. E10-012 · PROMPT 1.2
4. E5-502 y E5-503 (`MARKET_EXCLUDE_SEED=false`)
5. E10-013 · PROMPT 1.3
6. E5-504: `MARKET_EXCLUDE_SEED=true` recién cuando ese trim/año tenga ≥5 listings que no son seed
7. E10-014 · PROMPT 1.4

Después E10-020 (fase 2) → E10-030 (fase 3) → E10-040 (fase 4).

---

## Anti-patrones (decile esto a Cursor si se desvía)

```
NO hagas esto:
- Scrapear HTML de ML si la API alcanza
- Meter SQL de listings dentro de delivery/http
- Setear trim_id dentro del mapper de ML (debe ir por aliases)
- Mezclar CCA y ML en el mismo flag sin separación clara
- Expandir a 30 marcas antes de que ML + aliases funcionen
- Guardar PDFs completos de ACARA/CCA en el repo o en S3 público
```

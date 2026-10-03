# Mercado Libre API (MLA) — notas

Investigación. No hay cliente HTTP en este repo. No es asesoramiento legal. Ver [SOURCES.md](SOURCES.md): E5-501 está fuera de consideración; `use_ml_api = later` es nota histórica y no bloquea un cliente.

Probado el 2026-09-27: categoría **200** sin token; search **403** sin token.

## Registrar la app

1. [developers.mercadolibre.com.ar](https://developers.mercadolibre.com.ar) — país Argentina.
2. Crear aplicación ([guía](https://developers.mercadolibre.com.ar/en_us/en_us/register-your-application)): nombre, Redirect URI (HTTPS; en local puede ser de prueba), aceptar T&C.
3. Guardar Client ID y Client Secret. No commitearlos.
4. Un usuario de ML tiene que autorizar la app (OAuth). El token queda ligado a **esa cuenta**, no al marketplace entero.

Auth:

```text
https://auth.mercadolibre.com.ar/authorization?response_type=code&client_id=$APP_ID&redirect_uri=$REDIRECT_URI
```

```bash
curl -X POST \
  -H 'accept: application/json' \
  -H 'content-type: application/x-www-form-urlencoded' \
  'https://api.mercadolibre.com/oauth/token' \
  -d 'grant_type=authorization_code' \
  -d 'client_id=$APP_ID' \
  -d 'client_secret=$SECRET_KEY' \
  -d 'code=$SERVER_GENERATED_AUTHORIZATION_CODE' \
  -d 'redirect_uri=$REDIRECT_URI'
```

Refresh: `grant_type=refresh_token`. Token **solo** en header `Authorization: Bearer …`, nunca en query.

Docs: [autenticación](https://developers.mercadolibre.com.ar/en_us/categories-and-attributes/authentication-and-authorization).

## Categoría autos

`MLA1744` sigue vigente.

`GET https://api.mercadolibre.com/categories/MLA1744` (sin token, 200):

- `name`: Autos y Camionetas
- `settings.status`: enabled
- `settings.catalog_domain`: `MLA-CARS_AND_VANS`
- padre: `MLA1743` Autos, Motos y Otros

Atributos de la categoría: `GET /categories/MLA1744/attributes`  
Dominio: [CARS_AND_VANS](https://developers.mercadolibre.com.ar/en_us/en_us/domains-products-and-attributes-vehicle-accessories).

## Search

```text
GET https://api.mercadolibre.com/sites/MLA/search?category=MLA1744
```

Query útiles: `q=`, `condition=used`, `limit=`, `offset=`.  
Item: `GET /items/$ITEM_ID`.  
Search por vendedor (recurso privado): `GET /users/$USER_ID/items/search`.

Docs: [items & searches](https://developers.mercadolibre.com.ar/en_us/categories-and-attributes/items-and-searches).

## ¿Hace falta access token?

| Recurso | Sin token (2026-09-27) |
|---------|------------------------|
| `GET /categories/MLA1744` | 200 |
| `GET /sites/MLA/search?category=MLA1744` | **403** `forbidden` |

Desde ~2025 la search general no es anónima. Con Bearer, el token es de un usuario que autorizó la app: no asumas que podés indexar todo MLA. Guardar avisos ajenos para mediana sigue **Gris**.

## Campos → listing nuestro

`source` futuro: `mercadolibre` (aliases en migración 009).

| ML (`results[]` o `/items`) | Nosotros |
|-----------------------------|----------|
| `id` | `external_id` |
| `title` | `raw_title` |
| `price` | `price` |
| `currency_id` | `currency` |
| `permalink` | `url` |
| `attributes` id `VEHICLE_YEAR` | `year` |
| `attributes` id `KILOMETERS` | `km` |
| `attributes` `BRAND` / `MODEL` / `TRIM` o `SHORT_VERSION` | matching |
| `address` / `seller_address` | `location` |
| `condition` | filtro used |

Search no devolvió `results` sin token. Forma típica (docs ML, no captura de hoy):

```json
{
  "id": "MLA123",
  "title": "Toyota Corolla XEI 2.0 CVT",
  "price": 24800000,
  "currency_id": "ARS",
  "permalink": "https://auto.mercadolibre.com.ar/MLA-123-…",
  "condition": "used",
  "address": { "city_name": "Rosario", "state_name": "Santa Fe" },
  "attributes": [
    { "id": "VEHICLE_YEAR", "value_name": "2019" },
    { "id": "KILOMETERS", "value_name": "101000 km" },
    { "id": "BRAND", "value_name": "Toyota" },
    { "id": "MODEL", "value_name": "Corolla" },
    { "id": "TRIM", "value_name": "XEi 2.0 CVT" }
  ]
}
```

`title` igual entra a `normalize.Title` + `Resolve` (match exacto). “usado” / km en el título no pegan alias.

## Rate limits

No hay un número público único. 429 `too_many_requests` / `local_rate_limited`. El objeto app puede traer `max_requests_per_hour` (ej. 18000). Orientación interna: ~30–60/min endpoints sensibles, ~100–200/min generales. Search más estricto que GET item. Exceso → bloqueo `EXCESSIVE_API_CALL`. Token en header, no en URL.

## Curl

Categoría (funciona hoy, sin token):

```bash
curl -sS 'https://api.mercadolibre.com/categories/MLA1744'
```

Search sin token (hoy = 403):

```bash
curl -sS 'https://api.mercadolibre.com/sites/MLA/search?category=MLA1744&limit=1'
# {"message":"forbidden","error":"forbidden","status":403}
```

Search con token (redactar el Bearer; no commitear):

```bash
curl -sS -H 'Authorization: Bearer $ML_ACCESS_TOKEN' \
  'https://api.mercadolibre.com/sites/MLA/search?category=MLA1744&q=corolla&limit=2'
```

El package `internal/infrastructure/ingestion/mercadolibre` hace Search y mapea ítems. El worker sigue siendo archivo (`INGEST_FILE`) hasta E10-013.

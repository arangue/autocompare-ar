CREATE TABLE IF NOT EXISTS vehicle_listings (
    id BIGSERIAL PRIMARY KEY,
    source  TEXT NOT NULL,
    external_id TEXT NOT NULL,
    trim_id INT NOT NULL REFERENCES trims(id),
    year INT NOT NULL,
    km INT,
    price BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'ARS',
    location TEXT,
    url TEXT,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active        BOOLEAN     NOT NULL DEFAULT TRUE,

    UNIQUE (source, external_id)
);

CREATE INDEX IF NOT EXISTS vehicle_listings_trim_year_active_idx
    ON vehicle_listings (trim_id, year)
    WHERE active;
CREATE INDEX IF NOT EXISTS vehicle_listings_price_active_idx
    ON vehicle_listings (price)
    WHERE active;

INSERT INTO vehicle_listings (source, external_id, trim_id, year, km, price, currency, location, url)
SELECT 'seed', v.external_id, t.id, v.year, v.km, v.price, 'ARS', v.location, v.url
FROM trims t
JOIN generations g ON g.id = t.generation_id
JOIN (VALUES
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-01', 2019, 120000, 22500000, 'Córdoba', 'https://example.com/listings/seed-corolla-xei-2019-01'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-02', 2019,  98000, 23800000, 'Rosario', 'https://example.com/listings/seed-corolla-xei-2019-02'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-03', 2019, 105000, 24200000, 'Buenos Aires', 'https://example.com/listings/seed-corolla-xei-2019-03'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-04', 2019,  87000, 24800000, 'Mendoza', 'https://example.com/listings/seed-corolla-xei-2019-04'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-05', 2019, 112000, 25200000, 'La Plata', 'https://example.com/listings/seed-corolla-xei-2019-05'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-06', 2019,  76000, 25800000, 'Córdoba', 'https://example.com/listings/seed-corolla-xei-2019-06'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-07', 2019,  94000, 26000000, 'Rosario', 'https://example.com/listings/seed-corolla-xei-2019-07'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-08', 2019,  68000, 27100000, 'Buenos Aires', 'https://example.com/listings/seed-corolla-xei-2019-08'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-09', 2019,  55000, 28500000, 'Tucumán', 'https://example.com/listings/seed-corolla-xei-2019-09'),
    ('E210', 'XEi 2.0 CVT', 'seed-corolla-xei-2019-10', 2019,  72000, 29000000, 'Mar del Plata', 'https://example.com/listings/seed-corolla-xei-2019-10'),
    ('E210', 'XLi 1.8 CVT', 'seed-corolla-xli-2019-01', 2019, 135000, 19500000, 'Córdoba', 'https://example.com/listings/seed-corolla-xli-2019-01'),
    ('E210', 'XLi 1.8 CVT', 'seed-corolla-xli-2019-02', 2019, 118000, 20500000, 'Rosario', 'https://example.com/listings/seed-corolla-xli-2019-02'),
    ('E210', 'XLi 1.8 CVT', 'seed-corolla-xli-2019-03', 2019,  92000, 21800000, 'Buenos Aires', 'https://example.com/listings/seed-corolla-xli-2019-03'),
    ('Mk7', 'Comfortline', 'seed-golf-comfortline-2019-01', 2019, 145000, 18500000, 'Buenos Aires', 'https://example.com/listings/seed-golf-comfortline-2019-01'),
    ('Mk7', 'Comfortline', 'seed-golf-comfortline-2019-02', 2019, 128000, 19200000, 'Córdoba', 'https://example.com/listings/seed-golf-comfortline-2019-02')
) AS v(generation_name, trim_name, external_id, year, km, price, location, url)
    ON g.name = v.generation_name AND t.name = v.trim_name;

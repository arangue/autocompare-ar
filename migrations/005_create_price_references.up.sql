CREATE TABLE IF NOT EXISTS price_references (
    id          BIGSERIAL PRIMARY KEY,
    trim_id     INT         NOT NULL REFERENCES trims (id),
    year        INT         NOT NULL,
    source      TEXT        NOT NULL,
    kind        TEXT        NOT NULL CHECK (kind IN ('guide', 'fiscal', 'market')),
    price       BIGINT      NOT NULL,
    currency    TEXT        NOT NULL DEFAULT 'ARS',
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (trim_id, year, source)
);

INSERT INTO price_references (trim_id, year, source, kind, price, currency, observed_at)
SELECT t.id, v.year, v.source, v.kind, v.price, 'ARS', NOW()
FROM trims t
JOIN generations g ON g.id = t.generation_id
JOIN (VALUES
    ('E210', 'XEi 2.0 CVT', 2019, 'CCA',   'guide',  25500000),
    ('E210', 'XEi 2.0 CVT', 2019, 'ACARA', 'guide',  25800000),
    ('E210', 'XEi 2.0 CVT', 2019, 'DNRPA', 'fiscal', 24100000),
    ('E210', 'XLi 1.8 CVT', 2019, 'CCA',   'guide',  22800000),
    ('E210', 'XLi 1.8 CVT', 2019, 'ACARA', 'guide',  23100000),
    ('E210', 'XLi 1.8 CVT', 2019, 'DNRPA', 'fiscal', 21500000),
    ('XP210', 'XS 1.5 CVT', 2020, 'CCA',   'guide',  19800000),
    ('XP210', 'XS 1.5 CVT', 2020, 'ACARA', 'guide',  20100000),
    ('XP210', 'XS 1.5 CVT', 2020, 'DNRPA', 'fiscal', 18800000),
    ('Mk7', 'Comfortline', 2019, 'CCA',   'guide',  19200000),
    ('Mk7', 'Comfortline', 2019, 'ACARA', 'guide',  19500000),
    ('Mk7', 'Comfortline', 2019, 'DNRPA', 'fiscal', 18100000),
    ('1st gen', 'Drive 1.3', 2019, 'CCA',   'guide',  16500000),
    ('1st gen', 'Drive 1.3', 2019, 'ACARA', 'guide',  16800000),
    ('1st gen', 'Drive 1.3', 2019, 'DNRPA', 'fiscal', 15600000)
) AS v(generation_name, trim_name, year, source, kind, price)
    ON g.name = v.generation_name AND t.name = v.trim_name;

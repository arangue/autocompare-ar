CREATE TABLE IF NOT EXISTS brands (
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS models (
    id         SERIAL PRIMARY KEY,
    brand_id   INT         NOT NULL REFERENCES brands (id),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (brand_id, name)
);

CREATE TABLE IF NOT EXISTS generations (
    id         SERIAL PRIMARY KEY,
    model_id   INT         NOT NULL REFERENCES models (id),
    name       TEXT        NOT NULL,
    year_from  INT         NOT NULL,
    year_to    INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (model_id, name)
);

CREATE TABLE IF NOT EXISTS trims (
    id            SERIAL PRIMARY KEY,
    generation_id INT         NOT NULL REFERENCES generations (id),
    name          TEXT        NOT NULL,
    year_from     INT,
    year_to       INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (generation_id, name)
);

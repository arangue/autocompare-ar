-- Every price cell from a guide, whether or not it matches a catalog trim.
CREATE TABLE guide_prices (
    id          BIGSERIAL PRIMARY KEY,
    source      TEXT        NOT NULL,
    brand       TEXT        NOT NULL,
    model       TEXT        NOT NULL,
    trim_name   TEXT        NOT NULL,
    year        INT         NOT NULL,
    price       BIGINT      NOT NULL,
    currency    TEXT        NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    UNIQUE (source, brand, model, trim_name, year)
);

CREATE INDEX guide_prices_model_year ON guide_prices (model, year);

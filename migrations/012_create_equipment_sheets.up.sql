CREATE TABLE equipment_sheets (
    url                  TEXT           PRIMARY KEY,
    brand_name           TEXT           NOT NULL,
    model_name           TEXT           NOT NULL,
    title                TEXT           NOT NULL,
    year                 INT            NOT NULL,
    year_from            INT            NOT NULL,
    year_to              INT,
    engine               TEXT           NOT NULL DEFAULT '',
    displacement_cc      INT,
    horsepower           INT,
    transmission         TEXT           NOT NULL DEFAULT '',
    fuel                 TEXT           NOT NULL DEFAULT '',
    doors                INT,
    seats                INT,
    consumption_city     NUMERIC(4, 1),
    consumption_highway  NUMERIC(4, 1),
    features             JSONB          NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX equipment_sheets_model ON equipment_sheets (brand_name, model_name);

-- A catalog trim linked to one imported sheet for a calendar year.
-- Rows exist only when that link is unambiguous.
CREATE TABLE trim_equipment (
    trim_id   INT  NOT NULL REFERENCES trims (id) ON DELETE CASCADE,
    year      INT  NOT NULL,
    sheet_url TEXT NOT NULL REFERENCES equipment_sheets (url) ON DELETE CASCADE,
    PRIMARY KEY (trim_id, year)
);

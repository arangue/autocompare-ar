CREATE TABLE trim_aliases (
    id SERIAL PRIMARY KEY,
    source TEXT NOT NULL,
    raw_name TEXT NOT NULL,
    trim_id INT NOT NULL REFERENCES trims(id),
    year_from INT,
    year_to INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source, raw_name)
);

INSERT INTO trim_aliases (source, raw_name, trim_id)
SELECT 'file', v.raw, t.id
FROM trims t
JOIN generations g ON g.id = t.generation_id
JOIN (VALUES
    ('TOYOTA COROLLA XEI 2.0 CVT', 'XEi 2.0 CVT'),
    ('TOYOTA COROLLA XEI 2.0', 'XEi 2.0 CVT'),
    ('COROLLA XEI 2.0 CVT', 'XEi 2.0 CVT')
) AS v(raw, trim_name)
  ON t.name = v.trim_name AND g.name = 'E210';
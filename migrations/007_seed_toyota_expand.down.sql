DELETE FROM vehicle_features
WHERE trim_id IN (
    SELECT t.id FROM trims t
    JOIN generations g ON g.id = t.generation_id
    WHERE (g.name, t.name) IN (
        ('N80', 'SR 2.8 TDI 4x2'),
        ('N80', 'SRX 2.8 TDI 4x4'),
        ('AN160', 'SRX 2.8 TDI'),
        ('E12', 'XLS 1.5'),
        ('E12', 'Platinum 1.5'),
        ('XP210', 'XLS 1.5'),
        ('XA10', 'XEi 2.0'),
        ('XA10', 'SEG 2.0')
    )
);

DELETE FROM vehicle_specs
WHERE trim_id IN (
    SELECT t.id FROM trims t
    JOIN generations g ON g.id = t.generation_id
    WHERE (g.name, t.name) IN (
        ('N80', 'SR 2.8 TDI 4x2'),
        ('N80', 'SRX 2.8 TDI 4x4'),
        ('AN160', 'SRX 2.8 TDI'),
        ('E12', 'XLS 1.5'),
        ('E12', 'Platinum 1.5'),
        ('XP210', 'XLS 1.5'),
        ('XA10', 'XEi 2.0'),
        ('XA10', 'SEG 2.0')
    )
);

DELETE FROM trims
WHERE (generation_id, name) IN (
    SELECT g.id, v.trim_name
    FROM generations g
    JOIN (VALUES
        ('N80', 'SR 2.8 TDI 4x2'),
        ('N80', 'SRX 2.8 TDI 4x4'),
        ('AN160', 'SRX 2.8 TDI'),
        ('E12', 'XLS 1.5'),
        ('E12', 'Platinum 1.5'),
        ('XP210', 'XLS 1.5'),
        ('XA10', 'XEi 2.0'),
        ('XA10', 'SEG 2.0')
    ) AS v(generation_name, trim_name) ON g.name = v.generation_name
);

DELETE FROM generations
WHERE name IN ('N80', 'AN160', 'E12', 'XA10');

DELETE FROM models
WHERE name IN ('Hilux', 'SW4', 'Etios', 'Corolla Cross');

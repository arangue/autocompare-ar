DELETE FROM vehicle_features
WHERE trim_id IN (
    SELECT t.id FROM trims t
    JOIN generations g ON g.id = t.generation_id
    WHERE (g.name, t.name) IN (
        ('G5', 'Trendline'),
        ('AW', 'Comfortline'),
        ('2H', 'Comfortline'),
        ('2H', 'Highline'),
        ('226', 'Freedom 1.8'),
        ('331', 'Drive 1.3')
    )
);

DELETE FROM vehicle_specs
WHERE trim_id IN (
    SELECT t.id FROM trims t
    JOIN generations g ON g.id = t.generation_id
    WHERE (g.name, t.name) IN (
        ('G5', 'Trendline'),
        ('AW', 'Comfortline'),
        ('2H', 'Comfortline'),
        ('2H', 'Highline'),
        ('226', 'Freedom 1.8'),
        ('331', 'Drive 1.3')
    )
);

DELETE FROM trims
WHERE (generation_id, name) IN (
    SELECT g.id, v.trim_name
    FROM generations g
    JOIN (VALUES
        ('G5', 'Trendline'),
        ('AW', 'Comfortline'),
        ('2H', 'Comfortline'),
        ('2H', 'Highline'),
        ('226', 'Freedom 1.8'),
        ('331', 'Drive 1.3')
    ) AS v(generation_name, trim_name) ON g.name = v.generation_name
);

DELETE FROM generations
WHERE name IN ('G5', 'AW', '2H', '226', '331');

DELETE FROM models
WHERE name IN ('Gol', 'Polo', 'Amarok', 'Toro', 'Pulse');

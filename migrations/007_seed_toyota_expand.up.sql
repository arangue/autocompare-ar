INSERT INTO models (brand_id, name)
SELECT id, 'Hilux' FROM brands WHERE name = 'Toyota';

INSERT INTO models (brand_id, name)
SELECT id, 'SW4' FROM brands WHERE name = 'Toyota';

INSERT INTO models (brand_id, name)
SELECT id, 'Etios' FROM brands WHERE name = 'Toyota';

INSERT INTO models (brand_id, name)
SELECT id, 'Corolla Cross' FROM brands WHERE name = 'Toyota';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'N80', 2016, 2026 FROM models WHERE name = 'Hilux';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'AN160', 2016, 2026 FROM models WHERE name = 'SW4';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'E12', 2016, 2023 FROM models WHERE name = 'Etios';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'XA10', 2021, 2026 FROM models WHERE name = 'Corolla Cross';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'SR 2.8 TDI 4x2', 2016, 2026 FROM generations WHERE name = 'N80';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'SRX 2.8 TDI 4x4', 2016, 2026 FROM generations WHERE name = 'N80';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'SRX 2.8 TDI', 2016, 2026 FROM generations WHERE name = 'AN160';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'XLS 1.5', 2016, 2023 FROM generations WHERE name = 'E12';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Platinum 1.5', 2016, 2023 FROM generations WHERE name = 'E12';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'XLS 1.5', 2020, 2026 FROM generations WHERE name = 'XP210';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'XEi 2.0', 2021, 2026 FROM generations WHERE name = 'XA10';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'SEG 2.0', 2021, 2026 FROM generations WHERE name = 'XA10';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, v.engine, v.displacement_cc, v.horsepower, v.transmission, v.fuel, v.doors, v.seats, v.consumption_city, v.consumption_highway
FROM (VALUES
    ('N80', 'SR 2.8 TDI 4x2', '2.8 TDI', 2755, 177, 'manual', 'diesel', 4, 5, 10.5, 7.8),
    ('N80', 'SRX 2.8 TDI 4x4', '2.8 TDI', 2755, 177, 'automatic', 'diesel', 4, 5, 11.2, 8.4),
    ('AN160', 'SRX 2.8 TDI', '2.8 TDI', 2755, 177, 'automatic', 'diesel', 5, 7, 11.0, 8.2),
    ('E12', 'XLS 1.5', '1.5', 1496, 103, 'manual', 'nafta', 4, 5, 7.4, 5.6),
    ('E12', 'Platinum 1.5', '1.5', 1496, 103, 'automatic', 'nafta', 4, 5, 7.8, 5.9),
    ('XP210', 'XLS 1.5', '1.5', 1490, 107, 'CVT', 'nafta', 5, 5, 6.9, 5.3),
    ('XA10', 'XEi 2.0', '2.0', 1987, 170, 'CVT', 'nafta', 5, 5, 8.6, 6.4),
    ('XA10', 'SEG 2.0', '2.0', 1987, 170, 'CVT', 'nafta', 5, 5, 8.6, 6.4)
) AS v(generation_name, trim_name, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
JOIN generations g ON g.name = v.generation_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name;

INSERT INTO vehicle_features (trim_id, feature_id, value)
SELECT t.id, f.id, v.value
FROM (VALUES
    ('N80', 'SR 2.8 TDI 4x2', 'abs', 'true'),
    ('N80', 'SR 2.8 TDI 4x2', 'esp', 'true'),
    ('N80', 'SR 2.8 TDI 4x2', 'airbags', '2'),
    ('N80', 'SRX 2.8 TDI 4x4', 'abs', 'true'),
    ('N80', 'SRX 2.8 TDI 4x4', 'esp', 'true'),
    ('N80', 'SRX 2.8 TDI 4x4', 'airbags', '7'),
    ('N80', 'SRX 2.8 TDI 4x4', 'cruise_control', 'true'),
    ('AN160', 'SRX 2.8 TDI', 'abs', 'true'),
    ('AN160', 'SRX 2.8 TDI', 'esp', 'true'),
    ('AN160', 'SRX 2.8 TDI', 'airbags', '7'),
    ('AN160', 'SRX 2.8 TDI', 'isofix', 'true'),
    ('E12', 'XLS 1.5', 'abs', 'true'),
    ('E12', 'XLS 1.5', 'esp', 'false'),
    ('E12', 'XLS 1.5', 'airbags', '2'),
    ('E12', 'Platinum 1.5', 'abs', 'true'),
    ('E12', 'Platinum 1.5', 'esp', 'true'),
    ('E12', 'Platinum 1.5', 'airbags', '2'),
    ('XP210', 'XLS 1.5', 'abs', 'true'),
    ('XP210', 'XLS 1.5', 'esp', 'true'),
    ('XP210', 'XLS 1.5', 'airbags', '6'),
    ('XA10', 'XEi 2.0', 'abs', 'true'),
    ('XA10', 'XEi 2.0', 'esp', 'true'),
    ('XA10', 'XEi 2.0', 'airbags', '7'),
    ('XA10', 'SEG 2.0', 'abs', 'true'),
    ('XA10', 'SEG 2.0', 'esp', 'true'),
    ('XA10', 'SEG 2.0', 'airbags', '7'),
    ('XA10', 'SEG 2.0', 'cruise_control', 'true')
) AS v(generation_name, trim_name, code, value)
JOIN generations g ON g.name = v.generation_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name
JOIN features f ON f.code = v.code;

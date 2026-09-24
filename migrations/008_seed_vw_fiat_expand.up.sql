INSERT INTO models (brand_id, name)
SELECT id, 'Gol' FROM brands WHERE name = 'Volkswagen';

INSERT INTO models (brand_id, name)
SELECT id, 'Polo' FROM brands WHERE name = 'Volkswagen';

INSERT INTO models (brand_id, name)
SELECT id, 'Amarok' FROM brands WHERE name = 'Volkswagen';

INSERT INTO models (brand_id, name)
SELECT id, 'Toro' FROM brands WHERE name = 'Fiat';

INSERT INTO models (brand_id, name)
SELECT id, 'Pulse' FROM brands WHERE name = 'Fiat';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'G5', 2013, 2022 FROM models WHERE name = 'Gol';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'AW', 2018, 2026 FROM models WHERE name = 'Polo';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, '2H', 2010, 2022 FROM models WHERE name = 'Amarok';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, '226', 2016, 2026 FROM models WHERE name = 'Toro';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, '331', 2021, 2026 FROM models WHERE name = 'Pulse';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Trendline', 2013, 2022 FROM generations WHERE name = 'G5';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Comfortline', 2018, 2026 FROM generations WHERE name = 'AW';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Comfortline', 2010, 2022 FROM generations WHERE name = '2H';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Highline', 2010, 2022 FROM generations WHERE name = '2H';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Freedom 1.8', 2016, 2026 FROM generations WHERE name = '226';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Drive 1.3', 2021, 2026 FROM generations WHERE name = '331';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, v.engine, v.displacement_cc, v.horsepower, v.transmission, v.fuel, v.doors, v.seats, v.consumption_city, v.consumption_highway
FROM (VALUES
    ('G5', 'Trendline', '1.6', 1598, 101, 'manual', 'nafta', 5, 5, 8.1, 6.0),
    ('AW', 'Comfortline', '1.6 MSI', 1598, 110, 'automatic', 'nafta', 5, 5, 7.6, 5.5),
    ('2H', 'Comfortline', '2.0 TDI', 1968, 140, 'manual', 'diesel', 4, 5, 9.8, 7.4),
    ('2H', 'Highline', '3.0 V6 TDI', 2967, 258, 'automatic', 'diesel', 4, 5, 10.4, 8.0),
    ('226', 'Freedom 1.8', '1.8', 1747, 130, 'automatic', 'nafta', 4, 5, 9.2, 6.8),
    ('331', 'Drive 1.3', '1.3', 1332, 107, 'CVT', 'nafta', 5, 5, 6.8, 5.2)
) AS v(generation_name, trim_name, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
JOIN generations g ON g.name = v.generation_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name;

INSERT INTO vehicle_features (trim_id, feature_id, value)
SELECT t.id, f.id, v.value
FROM (VALUES
    ('G5', 'Trendline', 'abs', 'true'),
    ('G5', 'Trendline', 'esp', 'false'),
    ('G5', 'Trendline', 'airbags', '2'),
    ('AW', 'Comfortline', 'abs', 'true'),
    ('AW', 'Comfortline', 'esp', 'true'),
    ('AW', 'Comfortline', 'airbags', '6'),
    ('2H', 'Comfortline', 'abs', 'true'),
    ('2H', 'Comfortline', 'esp', 'true'),
    ('2H', 'Comfortline', 'airbags', '2'),
    ('2H', 'Highline', 'abs', 'true'),
    ('2H', 'Highline', 'esp', 'true'),
    ('2H', 'Highline', 'airbags', '7'),
    ('226', 'Freedom 1.8', 'abs', 'true'),
    ('226', 'Freedom 1.8', 'esp', 'true'),
    ('226', 'Freedom 1.8', 'airbags', '2'),
    ('331', 'Drive 1.3', 'abs', 'true'),
    ('331', 'Drive 1.3', 'esp', 'true'),
    ('331', 'Drive 1.3', 'airbags', '6')
) AS v(generation_name, trim_name, code, value)
JOIN generations g ON g.name = v.generation_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name
JOIN features f ON f.code = v.code;

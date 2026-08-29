CREATE TABLE IF NOT EXISTS vehicle_specs (
    trim_id             INT            PRIMARY KEY REFERENCES trims (id),
    engine              TEXT           NOT NULL,
    displacement_cc     INT,
    horsepower          INT,
    transmission        TEXT           NOT NULL,
    fuel                TEXT           NOT NULL,
    doors               INT,
    seats               INT,
    consumption_city    NUMERIC(4, 1),
    consumption_highway NUMERIC(4, 1)
);

CREATE TABLE IF NOT EXISTS features (
    id       SERIAL PRIMARY KEY,
    code     TEXT NOT NULL UNIQUE,
    name     TEXT NOT NULL,
    category TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicle_features (
    trim_id    INT  NOT NULL REFERENCES trims (id),
    feature_id INT  NOT NULL REFERENCES features (id),
    value      TEXT NOT NULL,
    PRIMARY KEY (trim_id, feature_id)
);

INSERT INTO features (code, name, category) VALUES
    ('abs', 'ABS', 'safety'),
    ('esp', 'ESP', 'safety'),
    ('airbags', 'Airbags', 'safety'),
    ('isofix', 'ISOFIX', 'safety'),
    ('front_disc', 'Frenos delanteros a disco', 'brakes'),
    ('rear_disc', 'Frenos traseros a disco', 'brakes'),
    ('cruise_control', 'Control de crucero', 'comfort');

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, '2.0', 1987, 170, 'CVT', 'nafta', 4, 5, 8.5, 6.2
FROM trims t
JOIN generations g ON g.id = t.generation_id
WHERE g.name = 'E210' AND t.name = 'XEi 2.0 CVT';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, '1.8', 1798, 140, 'CVT', 'nafta', 4, 5, 7.8, 5.9
FROM trims t
JOIN generations g ON g.id = t.generation_id
WHERE g.name = 'E210' AND t.name = 'XLi 1.8 CVT';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, '1.5', 1490, 107, 'CVT', 'nafta', 4, 5, 6.8, 5.2
FROM trims t
JOIN generations g ON g.id = t.generation_id
WHERE g.name = 'XP210' AND t.name = 'XS 1.5 CVT';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, '1.4 TSI', 1395, 150, 'DSG', 'nafta', 5, 5, 7.2, 5.4
FROM trims t
JOIN generations g ON g.id = t.generation_id
WHERE g.name = 'Mk7' AND t.name = 'Comfortline';

INSERT INTO vehicle_specs (trim_id, engine, displacement_cc, horsepower, transmission, fuel, doors, seats, consumption_city, consumption_highway)
SELECT t.id, '1.3', 1332, 99, 'manual', 'nafta', 4, 5, 7.5, 5.8
FROM trims t
JOIN generations g ON g.id = t.generation_id
WHERE g.name = '1st gen' AND t.name = 'Drive 1.3';

INSERT INTO vehicle_features (trim_id, feature_id, value)
SELECT t.id, f.id, v.value
FROM (VALUES
    ('E210', 'XEi 2.0 CVT', 'abs', 'true'),
    ('E210', 'XEi 2.0 CVT', 'esp', 'true'),
    ('E210', 'XEi 2.0 CVT', 'airbags', '7'),
    ('E210', 'XEi 2.0 CVT', 'isofix', 'true'),
    ('E210', 'XEi 2.0 CVT', 'front_disc', 'true'),
    ('E210', 'XEi 2.0 CVT', 'rear_disc', 'true'),
    ('E210', 'XEi 2.0 CVT', 'cruise_control', 'true'),
    ('E210', 'XLi 1.8 CVT', 'abs', 'true'),
    ('E210', 'XLi 1.8 CVT', 'esp', 'true'),
    ('E210', 'XLi 1.8 CVT', 'airbags', '2'),
    ('E210', 'XLi 1.8 CVT', 'isofix', 'true'),
    ('E210', 'XLi 1.8 CVT', 'front_disc', 'true'),
    ('E210', 'XLi 1.8 CVT', 'rear_disc', 'false'),
    ('XP210', 'XS 1.5 CVT', 'abs', 'true'),
    ('XP210', 'XS 1.5 CVT', 'esp', 'true'),
    ('XP210', 'XS 1.5 CVT', 'airbags', '6'),
    ('XP210', 'XS 1.5 CVT', 'isofix', 'true'),
    ('XP210', 'XS 1.5 CVT', 'front_disc', 'true'),
    ('XP210', 'XS 1.5 CVT', 'rear_disc', 'false'),
    ('Mk7', 'Comfortline', 'abs', 'true'),
    ('Mk7', 'Comfortline', 'esp', 'true'),
    ('Mk7', 'Comfortline', 'airbags', '6'),
    ('Mk7', 'Comfortline', 'isofix', 'true'),
    ('Mk7', 'Comfortline', 'front_disc', 'true'),
    ('Mk7', 'Comfortline', 'rear_disc', 'true'),
    ('Mk7', 'Comfortline', 'cruise_control', 'true'),
    ('1st gen', 'Drive 1.3', 'abs', 'true'),
    ('1st gen', 'Drive 1.3', 'esp', 'false'),
    ('1st gen', 'Drive 1.3', 'airbags', '2'),
    ('1st gen', 'Drive 1.3', 'isofix', 'true'),
    ('1st gen', 'Drive 1.3', 'front_disc', 'true'),
    ('1st gen', 'Drive 1.3', 'rear_disc', 'false')
) AS v(generation_name, trim_name, code, value)
JOIN generations g ON g.name = v.generation_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name
JOIN features f ON f.code = v.code;

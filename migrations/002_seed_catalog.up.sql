INSERT INTO brands (name) VALUES
    ('Toyota'),
    ('Volkswagen'),
    ('Fiat');

INSERT INTO models (brand_id, name)
SELECT id, 'Corolla' FROM brands WHERE name = 'Toyota';

INSERT INTO models (brand_id, name)
SELECT id, 'Golf' FROM brands WHERE name = 'Volkswagen';

INSERT INTO models (brand_id, name)
SELECT id, 'Cronos' FROM brands WHERE name = 'Fiat';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'E210', 2019, 2026 FROM models WHERE name = 'Corolla';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, 'Mk7', 2014, 2020 FROM models WHERE name = 'Golf';

INSERT INTO generations (model_id, name, year_from, year_to)
SELECT id, '1st gen', 2018, 2026 FROM models WHERE name = 'Cronos';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'XEi 2.0 CVT', 2019, 2026 FROM generations WHERE name = 'E210';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'XLi 1.8 CVT', 2019, 2026 FROM generations WHERE name = 'E210';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Comfortline', 2014, 2020 FROM generations WHERE name = 'Mk7';

INSERT INTO trims (generation_id, name, year_from, year_to)
SELECT id, 'Drive 1.3', 2018, 2026 FROM generations WHERE name = '1st gen';

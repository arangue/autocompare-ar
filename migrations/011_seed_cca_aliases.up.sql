-- CCA list lines do not match the file or ACARA aliases (4P/5P, comma decimals, list year).
-- raw_name is normalize.Title of brand + model + trim.
INSERT INTO trim_aliases (source, raw_name, trim_id)
SELECT 'cca', v.raw, t.id
FROM (VALUES
    ('TOYOTA COROLLA 4P 2,0 XEI CVT 2024', 'XEi 2.0 CVT', 'E210'),
    ('VOLKSWAGEN GOLF 5P 1,4 TSI COMFORTLINE 2018', 'Comfortline', 'Mk7'),
    ('FIAT CRONOS 4P 1,3 DRIVE GSE', 'Drive 1.3', '1st gen')
) AS v(raw, trim_name, gen_name)
JOIN generations g ON g.name = v.gen_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name;

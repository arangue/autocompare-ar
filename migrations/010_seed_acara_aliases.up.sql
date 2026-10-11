-- Guide lines that do not match an existing file alias.
-- raw_name is normalize.Title of brand + model + trim, with (170cv)/(L23) removed.
INSERT INTO trim_aliases (source, raw_name, trim_id)
SELECT 'acara', v.raw, t.id
FROM (VALUES
    ('TOYOTA COROLLA 2.0 XEI CVT', 'XEi 2.0 CVT', 'E210'),
    ('VOLKSWAGEN GOLF 1.4 TSI COMFORTLINE MT', 'Comfortline', 'Mk7'),
    ('FIAT CRONOS 1.3 GSE DRIVE MT', 'Drive 1.3', '1st gen')
) AS v(raw, trim_name, gen_name)
JOIN generations g ON g.name = v.gen_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name;

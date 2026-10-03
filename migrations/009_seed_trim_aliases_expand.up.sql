-- Extra aliases for seeded trims. raw_name is already normalize.Title output
-- (UPPER, no tildes, collapsed spaces). Lookup is exact: source + raw_name.
-- 006 already has TOYOTA COROLLA XEI 2.0 CVT / TOYOTA COROLLA XEI 2.0 / COROLLA XEI 2.0 CVT.
INSERT INTO trim_aliases (source, raw_name, trim_id)
SELECT s.source, v.raw, t.id
FROM (VALUES
    ('TOYOTA COROLLA XEI', 'XEi 2.0 CVT', 'E210'),
    ('COROLLA XEI', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XEI CVT', 'XEi 2.0 CVT', 'E210'),
    ('COROLLA XEI 2.0', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XEI 2.0 2019', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XEI 2.0 CVT 2019', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA 2.0 XEI CVT', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA 2.0 XEI CVT 2019', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XEI PACK', 'XEi 2.0 CVT', 'E210'),
    ('COROLLA XEI PACK', 'XEi 2.0 CVT', 'E210'),
    ('XEI PACK', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XEI PACK 2019', 'XEi 2.0 CVT', 'E210'),
    ('TOYOTA COROLLA XLI 1.8 CVT', 'XLi 1.8 CVT', 'E210'),
    ('TOYOTA COROLLA XLI 1.8', 'XLi 1.8 CVT', 'E210'),
    ('TOYOTA COROLLA XLI', 'XLi 1.8 CVT', 'E210'),
    ('COROLLA XLI 1.8 CVT', 'XLi 1.8 CVT', 'E210'),
    ('COROLLA XLI', 'XLi 1.8 CVT', 'E210'),
    ('TOYOTA COROLLA XLI 1.8 CVT 2019', 'XLi 1.8 CVT', 'E210'),
    ('VOLKSWAGEN GOLF COMFORTLINE', 'Comfortline', 'Mk7'),
    ('VW GOLF COMFORTLINE', 'Comfortline', 'Mk7'),
    ('GOLF COMFORTLINE', 'Comfortline', 'Mk7'),
    ('VOLKSWAGEN GOLF COMFORTLINE 2019', 'Comfortline', 'Mk7'),
    ('VW GOLF COMFORTLINE 2019', 'Comfortline', 'Mk7'),
    ('FIAT CRONOS DRIVE 1.3', 'Drive 1.3', '1st gen'),
    ('FIAT CRONOS DRIVE', 'Drive 1.3', '1st gen'),
    ('CRONOS DRIVE 1.3', 'Drive 1.3', '1st gen'),
    ('CRONOS DRIVE', 'Drive 1.3', '1st gen'),
    ('FIAT CRONOS DRIVE 1.3 2019', 'Drive 1.3', '1st gen')
) AS v(raw, trim_name, gen_name)
JOIN generations g ON g.name = v.gen_name
JOIN trims t ON t.generation_id = g.id AND t.name = v.trim_name
CROSS JOIN (VALUES ('file'), ('mercadolibre')) AS s(source);

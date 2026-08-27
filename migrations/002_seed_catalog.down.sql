DELETE FROM trims
WHERE name IN ('XEi 2.0 CVT', 'XLi 1.8 CVT', 'XS 1.5 CVT', 'Comfortline', 'Drive 1.3');

DELETE FROM generations
WHERE name IN ('E210', 'XP210', 'Mk7', '1st gen');

DELETE FROM models
WHERE name IN ('Corolla', 'Yaris', 'Golf', 'Cronos');

DELETE FROM brands
WHERE name IN ('Toyota', 'Volkswagen', 'Fiat');

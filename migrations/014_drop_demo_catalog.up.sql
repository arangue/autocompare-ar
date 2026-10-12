-- The catalog is the imported file data. Drop the hand-written demo rows.
DELETE FROM vehicle_listings vl
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = vl.trim_id);

DELETE FROM vehicle_features vf
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = vf.trim_id);

DELETE FROM vehicle_specs vs
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = vs.trim_id);

DELETE FROM price_references pr
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = pr.trim_id);

DELETE FROM trim_aliases ta
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = ta.trim_id);

DELETE FROM trims t
WHERE NOT EXISTS (SELECT 1 FROM trim_equipment te WHERE te.trim_id = t.id);

DELETE FROM generations g
WHERE NOT EXISTS (SELECT 1 FROM trims t WHERE t.generation_id = g.id);

DELETE FROM models m
WHERE NOT EXISTS (SELECT 1 FROM generations g WHERE g.model_id = m.id);

DELETE FROM brands b
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.brand_id = b.id);

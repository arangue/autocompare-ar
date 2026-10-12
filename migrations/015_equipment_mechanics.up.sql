ALTER TABLE equipment_sheets
    ADD COLUMN consumption_mixed NUMERIC(4, 1),
    ADD COLUMN torque_nm         INT,
    ADD COLUMN cylinders         INT,
    ADD COLUMN top_speed_kmh    INT,
    ADD COLUMN traction          TEXT NOT NULL DEFAULT '';

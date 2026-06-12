CREATE TABLE garbage_schedules (
    location_id INTEGER PRIMARY KEY, -- Unique building ID from system
    date_zmieszane DATE,
    date_papier DATE,
    date_plastik DATE,
    date_szklo DATE,
    date_bio DATE,
    date_zielone DATE,
    date_bio_restauracyjne DATE,
    date_gabaryty DATE,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
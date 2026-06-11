CREATE TABLE user_locations (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    address_name TEXT NOT NULL
);

CREATE INDEX idx_user_locations_location_id ON user_locations(location_id);
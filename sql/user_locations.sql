CREATE TABLE user_locations (
    id SERIAL PRIMARY KEY,    
    chat_id BIGINT NOT NULL,
    location_id VARCHAR(50) NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    address_name TEXT NOT NULL
);

CREATE INDEX idx_user_locations_location_id ON user_locations(location_id);

CREATE TABLE user_notifications (
    user_location_id INT REFERENCES user_locations(id) ON DELETE CASCADE,
    notification_time VARCHAR(50) NOT NULL,
    PRIMARY KEY (user_location_id, notification_time)
);
--
-- User
--
CREATE TABLE user(
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    default_location_id TEXT NOT NULL,
    PRIMARY KEY(id)
);
--
-- User
--
CREATE TABLE device(
    id TEXT NOT NULL,
    mac TEXT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(mac)
);
--
-- EOF
--

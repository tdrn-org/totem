--
-- User
--
CREATE TABLE user(
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    default_location_id TEXT NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(name)
);
CREATE TABLE user_device(
    mac TEXT NOT NULL,
    name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY(mac),
    FOREIGN KEY(user_id) REFERENCES user(id)
);
--
-- Log
--
CREATE TABLE presence_log(
    user_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    active INTEGER NOT NULL,
    ip TEXT,
    name TEXT,
    timestamp INTEGER NOT NULL,
    PRIMARY KEY(user_id,location_id,active),
    FOREIGN KEY(user_id) REFERENCES user(id)
);
--
-- EOF
--

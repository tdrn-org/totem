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
-- EOF
--

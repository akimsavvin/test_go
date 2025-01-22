CREATE TABLE users
(
    id         uuid PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name       VARCHAR(255),
    email      VARCHAR(255)
);

CREATE INDEX ON users USING hash (email);
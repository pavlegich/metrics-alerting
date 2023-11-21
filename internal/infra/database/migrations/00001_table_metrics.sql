-- +goose Up
CREATE TABLE IF NOT EXISTS storage (
    id text PRIMARY KEY, 
    value text NOT NULL
);

-- +goose Down
DROP TABLE storage;
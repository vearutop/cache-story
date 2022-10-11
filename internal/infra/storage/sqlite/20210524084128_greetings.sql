-- +goose Up
-- +goose StatementBegin
CREATE TABLE greetings
(
    `id`         INTEGER PRIMARY KEY,
    `created_at` DATETIME               NOT NULL DEFAULT current_timestamp,
    `message`    VARCHAR(255)           NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `greetings`;
-- +goose StatementEnd

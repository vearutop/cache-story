-- +goose Up
-- +goose StatementBegin
CREATE TABLE `greetings`
(
    `id`         INTEGER AUTO_INCREMENT NOT NULL,
    `created_at` DATETIME               NOT NULL DEFAULT current_timestamp(),
    `message`    VARCHAR(255)           NOT NULL UNIQUE,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARACTER SET = utf8;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `greetings`;
-- +goose StatementEnd

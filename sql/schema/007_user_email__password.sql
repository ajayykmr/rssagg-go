-- +goose Up
alter table users
add email varchar(255) not null unique,
add password char(60);

-- +goose Down
alter table users
drop column email,
drop column password;
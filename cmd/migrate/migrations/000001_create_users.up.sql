CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE If Not EXISTS users(
id bigserial PRIMARY KEY,
email citext UNIQUE NOT NULL,
username varchar(255) UNIQUE NOT NULL,
password bytea NOT NULL,
created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
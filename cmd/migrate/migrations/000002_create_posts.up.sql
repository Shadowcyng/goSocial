
CREATE TABLE If Not EXISTS posts(
id bigserial PRIMARY KEY,
content text  NOT NULL,
title text  NOT NULL,
user_id bigint NOT NULL,
created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
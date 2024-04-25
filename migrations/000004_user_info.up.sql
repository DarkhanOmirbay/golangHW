CREATE TABLE IF NOT EXISTS user_info (
                                     id bigserial PRIMARY KEY,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    fname VARCHAR(255) NOT NULL,
    sname  VARCHAR(255) NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    user_role VARCHAR(255),
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
    );
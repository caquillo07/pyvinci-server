-- uuid v4 generator
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE user_record
(
    id         uuid primary key default uuid_generate_v4(),
    username   TEXT UNIQUE NOT NULL,
    password   TEXT        NOT NULL,
    created_at TIMESTAMP   NOT NULL,
    updated_at TIMESTAMP   NOT NULL
);

CREATE INDEX idx_user_username on user_record (username);

CREATE TABLE token
(
    id          uuid primary key                          default uuid_generate_v4(),
    user_record uuid REFERENCES user_record (id) NOT NULL,
    valid       BOOLEAN                          NOT NULL DEFAULT TRUE,
    token       TEXT UNIQUE                      NOT NULL,
    created_at  TIMESTAMP                        NOT NULL,
    updated_at  TIMESTAMP                        NOT NULL
);

CREATE INDEX idx_token_token on token (token);

CREATE TABLE project
(
    id          uuid primary key default uuid_generate_v4(),
    user_record uuid REFERENCES user_record (id) NOT NULL,
    keywords    TEXT ARRAY,
    created_at  TIMESTAMP                        NOT NULL,
    updated_at  TIMESTAMP                        NOT NULL
);

CREATE TABLE image
(
  id uuid primary key default uuid_generate_v4(),
  project_id uuid REFERENCES project (id) NOT NULL,
  url TEXT NOT NULL,
  labels_things TEXT ARRAY,
  labels_stuff TEXT ARRAY,
  masks_labels TEXT ARRAY,
  masks BYTEA,
  created_at  TIMESTAMP                        NOT NULL,
  updated_at  TIMESTAMP                        NOT NULL
);

CREATE TABLE jobs
(
    id uuid primary key default uuid_generate_v4(),
    project_id uuid REFERENCES project (id) NOT NULL,
    result_image_url  TEXT,
    status      TEXT NOT NULL,
    created_at  TIMESTAMP                        NOT NULL,
    updated_at  TIMESTAMP                        NOT NULL
);

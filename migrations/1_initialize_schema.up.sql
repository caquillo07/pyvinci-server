CREATE TABLE user_record (
  id         SERIAL PRIMARY KEY,
  uid        TEXT        NOT NULL,
  username   TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP   NOT NULL,
  updated_at TIMESTAMP   NOT NULL
);

CREATE INDEX idx_users_uid on user_record (uid);

CREATE TABLE token (
  id          SERIAL PRIMARY KEY,
  user_record INTEGER REFERENCES user_record (id) NOT NULL,
  valid       BOOLEAN                             NOT NULL DEFAULT TRUE,
  token       TEXT UNIQUE                         NOT NULL,
  created_at  TIMESTAMP                           NOT NULL,
  updated_at  TIMESTAMP                           NOT NULL
);

CREATE INDEX idx_token_token on token (token);



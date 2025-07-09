-- +goose Up
CREATE TABLE repo (
  owner text,
  name text,
  state jsonb NOT NULL,
  CONSTRAINT repo_pk PRIMARY KEY (owner, name)
);

CREATE TABLE repo_event (
  id uuid,
  owner text NOT NULL,
  name text NOT NULL,
  timestamp timestamptz NOT NULL,
  sequence int NOT NULL,
  data jsonb NOT NULL,
  state jsonb NOT NULL,
  CONSTRAINT repo_event_pk PRIMARY KEY (id),
  CONSTRAINT repo_event_fk_state FOREIGN KEY (owner, name) REFERENCES repo(owner, name)
);

-- +goose Down

DROP TABLE repo_event;
DROP TABLE repo;


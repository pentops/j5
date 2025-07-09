-- +goose Up

CREATE TABLE j5_version (
  owner text,
  repo text,
  version text,
  data jsonb NOT NULL,

  CONSTRAINT j5_version_pk PRIMARY KEY (owner, repo, version)
);

CREATE TABLE go_module_version (
  package_name text,
  version text,
  timestamp timestamptz NOT NULL,
  data jsonb NOT NULL,

  CONSTRAINT go_module_version_pk PRIMARY KEY (package_name, version)
);

-- +goose Down

DROP TABLE j5_version;
DROP TABLE go_module_version;

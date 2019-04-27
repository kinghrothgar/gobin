CREATE DATABASE gobin;
create table gobin.gob_metadata (
	id          STRING PRIMARY KEY,
	secret      STRING UNIQUE NOT NULL,
	encrypted   BOOL,
	create_date  TIMESTAMP,
	expire_date  TIMESTAMP,
	size        INT,
    filename     STRING,
	content_type STRING,
	owner_id     INT,
);

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE gobin.gob_metadata TO gobin;

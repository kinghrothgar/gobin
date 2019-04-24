CREATE DATABASE gobin;
create table gobin.gob_metadata (
	id          STRING PRIMARY KEY,
	auth_key     STRING UNIQUE NOT NULL,
	encrypted   BOOL,
	create_date  TIMESTAMP,
	expire_date  TIMESTAMP,
	size        INT,
	owner_id     INT,
	content_type STRING
);

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE gobin.gob_metadata TO gobin;

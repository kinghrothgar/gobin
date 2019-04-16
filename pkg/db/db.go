package db

import (
	"context"
	"errors"

	_ "github.com/jackc/pgx/pgtype"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kinghrothgar/gobin/pkg/gob"

	// This is needed for sqlx.Connect
	_ "github.com/lib/pq"
)

var db *sqlx.DB

// TODO How to not require an init to do this
func Connect(ctx context.Context, dataSourceName string) (*sqlx.DB, error) {
	newDB, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = newDB.PingContext(ctx); err != nil {
		return nil, err
	}
	db = newDB
	return db, err
}

// TODO is there any point to not exporting a var containing a pointer?
func DB() *sqlx.DB {
	return db
}

func InsertMetadata(meta *gob.Metadata) error {
	if db == nil {
		return errors.New("No DB connected")
	}
	q := "INSERT INTO gob_metadata (" +
		"id, auth_key, encrypted, create_date, " +
		"expire_date, size, owner_id, content_type)" +
		"VALUES(" +
		":id, :auth_key, :encrypted, :create_date, " +
		":expire_date, :size, :owner_id, :content_type)"
	_, err := db.NamedExec(q, meta)
	return err
}

func GetMetadata(id string, authKey string) *gob.Metadata {
	return nil
}

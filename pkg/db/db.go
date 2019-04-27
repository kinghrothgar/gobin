package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

// TODO should be using context for all queries

// TODO How to not require an init to do this
func Connect(ctx context.Context, dataSourceName string) (*DB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) InsertMetadata(meta *Metadata) error {
	if db == nil {
		return errors.New("no db connected")
	}
	q := "INSERT INTO gob_metadata (" +
		"id, auth_key, encrypted, create_date, " +
		"expire_date, size, owner_id, content_type, filename)" +
		"VALUES(" +
		":id, :auth_key, :encrypted, :create_date, " +
		":expire_date, :size, :owner_id, :content_type, :filename)"
	_, err := db.NamedExec(q, meta)
	return err
}

func (db *DB) GetMetadataByID(id string) (*Metadata, error) {
	if db == nil {
		return nil, errors.New("no db connected")
	}
	meta := NewMetadata()
	// TODO: should select specify the coloumns
	err := db.QueryRowx("SELECT * FROM gob_metadata WHERE id=$1", id).StructScan(meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (db *DB) GetMetadataByAuthKey(authKey string) (*Metadata, error) {
	if db == nil {
		return nil, errors.New("no db connected")
	}
	meta := NewMetadata()
	// TODO: should select specify the coloumns
	err := db.QueryRowx("SELECT * FROM gob_metadata WHERE auth_key=$1", authKey).StructScan(meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (db *DB) DeleteMetadataByAuthKey(authKey string) error {
	if db == nil {
		return errors.New("no db connected")
	}
	result, err := db.Exec("DELETE FROM gob_metadata WHERE auth_key=$1", authKey)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if numRows != 1 {
		return errors.New("failed to delete metatdata")
	}
	return nil
}

func (db *DB) UpdateMetadata(meta *Metadata) error {
	if db == nil {
		return errors.New("no db connected")
	}
	q := "UPDATE gob_metadata SET (" +
		"encrypted, create_date, expire_date, " +
		"size, owner_id, content_type, filename) = (" +
		":encrypted, :create_date, :expire_date, " +
		":size, :owner_id, :content_type, :filename) " +
		"WHERE id = :id"
	result, err := db.NamedExec(q, meta)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if numRows != 1 {
		return errors.New("failed to update metatdata")
	}
	return nil
}

// NewInsertedMetadata returns new *Metadata that has been successfully inserted into db
// TODO create an entirely new struct each time not efficient
// TODO atleast unset old struct?
func (db *DB) NewInsertedMetadata(tries int) (*Metadata, error) {
	var meta *Metadata
	for i := 0; i < tries; i++ {
		meta = NewMetadata()
		err := db.InsertMetadata(meta)
		if IsUniqueViolation(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return meta, nil
	}
	return nil, fmt.Errorf("failed to insert new metatdata in %d tries", tries)
}

func IsUniqueViolation(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		if err.Code.Name() == "unique_violation" {
			return true
		}
	}
	return false
}

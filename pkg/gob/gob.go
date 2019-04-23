package gob

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/store"
)

var bucketName = "gobin-io-test"

// TODO review concurrency

// TODO does object dangle of upload not completed?
func Upload(ctx context.Context, reader io.Reader, encryptKey string) (*db.Metadata, error) {
	meta, err := db.NewInsertedMetadata(3)
	obj, err := store.NewObject(ctx, bucketName, meta.ID)
	if err != nil {
		return nil, err
	}
	// TODO: should I be checking if it exists or let metadata be master
	if exists, err := obj.Exists(ctx); err != nil {
		return nil, err
	} else if exists {
		return nil, fmt.Errorf("store %s already exists", meta.ID)
	}
	// TODO how to set salt?
	if encryptKey != "" {
		meta.Encrypted = true
		obj.Key(encryptKey, "saltsaltsalt")
	}
	w, err := obj.NewWriter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get store %s writer: %v", meta.ID, err)
	}

	// Sniff content type
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = reader.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read first 512 bytes of %s: %v", meta.ID, err)
	}
	meta.SetContentType(buffer)

	// Write to storage
	w.Write(buffer)
	meta.Size, err = io.Copy(w, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy %s to store: %v", meta.ID, err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close %s store: %v", meta.ID, err)
	}
	if err := db.UpdateMetadata(meta); err != nil {
		return meta, fmt.Errorf("failed to update %s metadata: %v", meta.ID, err)
	}
	return meta, nil
}

func Download(ctx context.Context, writer io.Writer, id string, encryptKey string) (*db.Metadata, error) {
	meta, err := db.GetMetadataByID(id)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return nil, err
	}
	if meta.ExpireDate.Valid && time.Now().After(meta.ExpireDate.Time) {
		return nil, fmt.Errorf("%s expired", meta.ID)
	}
	obj, err := store.NewObject(ctx, bucketName, meta.ID)
	if err != nil {
		return nil, err
	}
	if exists, err := obj.Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("store %s does not exist", meta.ID)
	}
	// TODO how to set salt?
	if meta.Encrypted && encryptKey == "" {
		// TODO probably should return typed error
		return nil, fmt.Errorf("store %s requires encrypt key", meta.ID)
	} else if meta.Encrypted {
		obj.Key(encryptKey, "saltsaltsalt")
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		// TODO return typed error for CustomerEncryptionKeyIsIncorrect 400
		return nil, fmt.Errorf("failed to get store %s reader: %v", meta.ID, err)
	}
	if _, err := io.Copy(writer, r); err != nil {
		return nil, fmt.Errorf("failed to copy %s from store: %v", meta.ID, err)
	}
	if err := r.Close(); err != nil {
		return nil, fmt.Errorf("failed to close %s store: %v", meta.ID, err)
	}
	return meta, nil
}

func Expire(ctx context.Context, authKey string) (*db.Metadata, error) {
	meta, err := db.GetMetadataByAuthKey(authKey)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return nil, err
	}
	if meta.ExpireDate.Valid && time.Now().After(meta.ExpireDate.Time) {
		return nil, fmt.Errorf("%s expired", meta.ID)
	}
	meta.SetExpireDate(time.Now())
	if err = db.UpdateMetadata(meta); err != nil {
		return nil, fmt.Errorf("failed to expire %s gob: %v", meta.ID, err)
	}
	return meta, nil
}

func Delete(ctx context.Context, authKey string) error {
	meta, err := db.GetMetadataByAuthKey(authKey)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return err
	}
	obj, err := store.NewObject(ctx, bucketName, meta.ID)

	if err != nil {
		return err
	}
	if exists, err := obj.Exists(ctx); err != nil {
		return err
	} else if exists {
		obj.Delete(ctx)
	}
	return db.DeleteMetadataByAuthKey(authKey)
}

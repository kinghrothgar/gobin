package gob

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/store"
	"github.com/levenlabs/errctx"
)

// TODO review concurrency
// TODO review ctx

// TODO how to pass this in
var bucketName = "gobin-io-test"

type Gob struct {
	ctx context.Context
	db  *db.DB
}

func NewGob(ctx context.Context, db *db.DB) *Gob {
	return &Gob{ctx, db}
}

// cleanUpMetadata deletes metadata and passes nil, error
func (gob *Gob) failedUploadHelper(secret string, err error) (*db.Metadata, error) {
	gob.db.DeleteMetadataBySecret(secret)
	return nil, err
}

// TODO does object dangle of upload not completed?
func (gob *Gob) Upload(reader io.Reader, encryptKey string, filename string) (*db.Metadata, error) {
	meta, err := gob.db.NewInsertedMetadata(3)
	if err != nil {
		return nil, err
	}
	obj, err := store.NewObject(gob.ctx, bucketName, meta.ID)
	if err != nil {
		return gob.failedUploadHelper(meta.Secret, err)
	}
	// TODO: should I be checking if it exists or let metadata be master
	if exists, err := obj.Exists(gob.ctx); err != nil {
		return gob.failedUploadHelper(meta.Secret, err)
	} else if exists {
		err := errctx.Mark(fmt.Errorf("store %s already exists", meta.ID))
		return gob.failedUploadHelper(meta.Secret, err)
	}
	// TODO how to set salt?
	if encryptKey != "" {
		meta.Encrypted = true
		obj.Key(encryptKey, "saltsaltsalt")
	}

	// Sniff content type
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	bytesRead, err := reader.Read(buffer)
	if err != nil {
		return gob.failedUploadHelper(meta.Secret, errctx.Mark(err))
	}
	meta.SetContentType(buffer[:bytesRead])

	// Write to storage
	w := obj.NewWriter(gob.ctx)
	w.Write(buffer)
	meta.Size, err = store.Copy(gob.ctx, w, reader)
	if err != nil {
		return gob.failedUploadHelper(meta.Secret, errctx.Mark(err))
	}
	meta.Size += int64(bytesRead)
	if err := w.Close(); err != nil {
		// TODO this could leave a dangling storage obj?
		err = errctx.Mark(fmt.Errorf("failed to close %s store: %v", meta.ID, err))
		return gob.failedUploadHelper(meta.Secret, err)
	}

	// Update metadata
	meta.SetFilename(filename)
	if err := gob.db.UpdateMetadata(meta); err != nil {
		err = errctx.Mark(fmt.Errorf("failed to update %s metadata: %v", meta.ID, err))
		return nil, err
	}
	return meta, nil
}

func (gob *Gob) GetMetadata(id string) (*db.Metadata, error) {
	meta, err := gob.db.GetMetadataByID(id)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return nil, err
	}
	// TODO probably should return typed error if Expired
	if meta.ExpireDate.Valid && time.Now().After(meta.ExpireDate.Time) {
		return nil, fmt.Errorf("%s expired", meta.ID)
	}
	return meta, nil
}

func (gob *Gob) Download(w io.Writer, meta *db.Metadata, encryptKey string) error {
	obj, err := store.NewObject(gob.ctx, bucketName, meta.ID)
	if err != nil {
		return err
	}
	if exists, err := obj.Exists(gob.ctx); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("store %s does not exist", meta.ID)
	}
	// TODO how to set salt?
	if meta.Encrypted && encryptKey == "" {
		// TODO probably should return typed error
		return fmt.Errorf("store %s requires encrypt key", meta.ID)
	} else if meta.Encrypted {
		obj.Key(encryptKey, "saltsaltsalt")
	}

	r, err := obj.NewReader(gob.ctx)
	if err != nil {
		// TODO return typed error for CustomerEncryptionKeyIsIncorrect 400
		return errctx.Mark(fmt.Errorf("failed to get store %s reader: %v", meta.ID, err))
	}
	if _, err := store.Copy(gob.ctx, w, r); err != nil {
		return errctx.Mark(fmt.Errorf("failed to copy %s from store: %v", meta.ID, err))
	}
	if err := r.Close(); err != nil {
		return errctx.Mark(fmt.Errorf("failed to close %s store: %v", meta.ID, err))
	}
	return nil
}

func (gob *Gob) Expire(secret string) (*db.Metadata, error) {
	meta, err := gob.db.GetMetadataBySecret(secret)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return nil, err
	}
	if meta.ExpireDate.Valid && time.Now().After(meta.ExpireDate.Time) {
		return nil, fmt.Errorf("%s expired", meta.ID)
	}
	meta.SetExpireDate(time.Now())
	if err = gob.db.UpdateMetadata(meta); err != nil {
		return nil, fmt.Errorf("failed to expire %s gob: %v", meta.ID, err)
	}
	return meta, nil
}

func (gob *Gob) Delete(secret string) error {
	meta, err := gob.db.GetMetadataBySecret(secret)
	// TODO probably should return typed error if id does not exist
	if err != nil {
		return err
	}
	obj, err := store.NewObject(gob.ctx, bucketName, meta.ID)

	if err != nil {
		return err
	}
	if exists, err := obj.Exists(gob.ctx); err != nil {
		return err
	} else if exists {
		obj.Delete(gob.ctx)
	}
	return gob.db.DeleteMetadataBySecret(secret)
}

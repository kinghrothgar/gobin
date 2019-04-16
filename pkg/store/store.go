package store

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/DataDog/zstd"
	"golang.org/x/crypto/scrypt"
)

type Writer struct {
	writer  io.WriteCloser
	closers []io.WriteCloser
}

type Reader struct {
	reader  io.ReadCloser
	closers []io.ReadCloser
}

// Object wraps google storage.ObjectHandle
type Object struct {
	*storage.ObjectHandle
}

// Write writes a compressed form of p to the underlying io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

// Close all writers
func (w *Writer) Close() error {
	for _, closer := range w.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Reade reades a compressed form of p to the underlying io.Reader.
func (w *Reader) Read(p []byte) (int, error) {
	return w.reader.Read(p)
}

// Close all readers
func (w *Reader) Close() error {
	for _, closer := range w.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (obj *Object) NewWriter(ctx context.Context) (*Writer, error) {
	w := obj.ObjectHandle.NewWriter(ctx)
	zw := zstd.NewWriter(w)
	return &Writer{
		writer:  zw,
		closers: []io.WriteCloser{zw, w},
	}, nil
}

func (obj *Object) NewReader(ctx context.Context) (*Reader, error) {
	r, err := obj.ObjectHandle.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	zr := zstd.NewReader(r)
	return &Reader{
		reader:  zr,
		closers: []io.ReadCloser{zr, r},
	}, nil

}

func (obj *Object) Key(pass string, salt string) error {
	key, err := NewKey(pass, salt)
	if err != nil {
		return err
	}
	obj.ObjectHandle = obj.ObjectHandle.Key(key)
	return nil
}

func (obj *Object) Metadata(ctx context.Context) (map[string]string, error) {
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return attrs.Metadata, nil
}

func (obj *Object) Exists(ctx context.Context) (bool, error) {
	// TODO: doesn't seem like a easy way to check objects existance
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (obj *Object) UpdateMetadata(ctx context.Context, meta map[string]string) (map[string]string, error) {
	updateAttrs := storage.ObjectAttrsToUpdate{
		Metadata: meta,
	}
	attrs, err := obj.Update(ctx, updateAttrs)
	if err != nil {
		return nil, err
	}
	return attrs.Metadata, nil
}

func NewObject(ctx context.Context, bucketName string, path string) (*Object, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	bkt := client.Bucket(bucketName)
	return &Object{
		bkt.Object(path),
	}, nil
}

func NewKey(pass string, salt string) ([]byte, error) {
	return scrypt.Key([]byte(pass), []byte(salt), 32768, 8, 1, 32)
}

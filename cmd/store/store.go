package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/kinghrothgar/gobin/pkg/store"
)

func main() {
	test()
	log.Println("done")
}

var bucketName = "gobin-io-test"

func test() {
	ctx := context.Background()
	obj, err := store.NewObject(ctx, bucketName, "data")
	if err != nil {
		log.Fatalf("failed to get new object: %v", err)
	}
	_, err = obj.Exists(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	obj.Key("asdfasdfasdf", "saltsaltsalt")
	w, err := obj.NewWriter(ctx)
	if err != nil {
		log.Fatalf("failed to get object writer: %v", err)
	}
	log.Println("starting to copy")
	b, err := io.Copy(w, os.Stdin)
	if err != nil {
		log.Fatalf("failed to read from stdin to object: %v", err)
	} else {
		log.Printf("Read in %d", b)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("failed to close to object: %v", err)
	}
	log.Println("done closing writer")

	// Read it back.
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("failed to get object reader: %v", err)
	}
	f, err := os.OpenFile("out.tar", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer r.Close()
	defer f.Close()
	log.Println("starting to copy")
	if _, err := io.Copy(f, r); err != nil {
		log.Fatalf("failed to copy to file: %v", err)
	}
	if meta, err := obj.Metadata(ctx); err != nil {
		log.Fatalf("failed to get object metadata: %v", err)
	} else {
		log.Printf("metadata is %+v", meta)
	}
	meta := map[string]string{
		"rawSize": strconv.FormatInt(b, 10),
	}
	if meta, err := obj.UpdateMetadata(ctx, meta); err != nil {
		log.Fatalf("failed to update object metadata: %v", err)
	} else {
		log.Printf("metadata is %+v", meta)
	}
	log.Println("deleting object")
	if err := obj.Delete(ctx); err != nil {
		log.Fatalf("failed to delete object: %v", err)
	}
}

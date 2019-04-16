package main

import (
	"context"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/kinghrothgar/gobin/pkg/store"
)

func main() {
	//rand.Seed(68421)
	//fmt.Println(RandStringRunes(6))
	//fmt.Println(RandStringRunes(6))
	object()
	log.Println("done")
}

var letterRunes = []rune("ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")

var bucketName = "gobin-io-test"

//func RandStringRunes(n int) string {
//	b := make([]rune, n)
//	for i := range b {
//		b[i] = letterRunes[rand.Intn(len(letterRunes))]
//	}
//	return string(b)
//}

func bucket() {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	//projectID := "gobin-io"

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Sets the name for the new bucket.
	bucketName := "gobin-io-test"

	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	// Creates the new bucket.
	if attr, err := bucket.Attrs(ctx); err != nil {
		log.Fatalf("Failed to get attributes of bucket: %v", err)
	} else {
		log.Printf("%v", attr)
	}
}

func object() {
	ctx := context.Background()
	obj, err := store.NewObject(ctx, bucketName, "data")
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
	log.Println("deleting object")

}

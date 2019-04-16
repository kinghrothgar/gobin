package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gob"
	"github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	_, err := db.Connect(ctx, "host=127.0.0.1 port=26257 user=gobin dbname=gobin sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	meta := gob.NewMetadata()
	log.Println(meta.ExpireDate)
	err = db.InsertMetadata(meta)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			fmt.Println("pq error:", err.Code.Name())
		}
	}
}

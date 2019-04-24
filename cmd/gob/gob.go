package main

import (
	"context"
	"log"
	"os"

	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gob"
)

func main() {
	ctx := context.Background()
	db, err := db.Connect(ctx, "host=127.0.0.1 port=26257 user=gobin dbname=gobin sslmode=disable")
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	g := gob.NewGob(ctx, db)
	_, err = g.Upload(os.Stdin, "asdf")
	f, err := os.OpenFile("E.coli.down", os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	_, err = g.Download(f, "7kerps", "")
	if err != nil {
		log.Fatal(err)
	}
}

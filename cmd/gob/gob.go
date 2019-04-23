package main

import (
	"context"
	"log"

	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gob"
)

func main() {
	ctx := context.Background()
	_, err := db.Connect(ctx, "host=127.0.0.1 port=26257 user=gobin dbname=gobin sslmode=disable")
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	//_, err = gob.Upload(ctx, os.Stdin, "asdf")
	//f, err := os.OpenFile("E.coli.down", os.O_CREATE|os.O_WRONLY, 0666)
	//defer f.Close()
	//meta, err = gob.Download(ctx, f, "7kerps", "")
	//if err != nil {
	//	log.Fatal(err)
	//}
	for _, authKey := range [...]string{"rMqMbiKQZfnSFh9h", "2SrEU3rdYYxK9CuQ", "7x3vwUmmQNLGpEIb"} {
		err := gob.Delete(ctx, authKey)
		if err != nil {
			log.Fatal(err)
		}
	}
}

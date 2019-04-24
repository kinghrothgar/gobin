package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gobin"
	"github.com/levenlabs/go-llog"
)

func main() {
	var shutDownGracePeriod = time.Second * 60

	ctx := context.Background()
	llog.SetLevelFromString("DEBUG")

	db, err := db.Connect(ctx, "host=127.0.0.1 port=26257 user=gobin dbname=gobin sslmode=disable")
	if err != nil {
		llog.Fatal("failed to connect to db", llog.KV{"err": err})
	}

	r := mux.NewRouter()
	r.Handle("/", gobin.PostGobHandler(db)).Methods("POST")
	r.Handle("/{id}", gobin.GetGobHandler(db)).Methods("GET")
	//r.HandleFunc("/articles/{category}/", ArticlesCategoryHandler)
	//r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
	//mux.Get("/", http.HandlerFunc(handler.GetRoot))
	//mux.Get("/:uid", http.HandlerFunc(handler.GetGob))
	//mux.Get("/delete/:token", http.HandlerFunc(handler.DelGob))
	//mux.Post("/append/:token", http.HandlerFunc(handler.AppendGob))
	//mux.Get("/horde/:horde", http.HandlerFunc(handler.GetHorde))
	//mux.Get("/new/gob", http.HandlerFunc(handler.GetForm))
	//mux.Post("/", http.HandlerFunc(handler.PostGob))
	//// TODO: Should I post to /horde/:horde
	//mux.Post("/:horde", http.HandlerFunc(handler.PostHordeGob))

	srv := &http.Server{
		Addr: "127.0.0.1:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		llog.Info("gobin is listening")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				llog.Fatal("failed to listen and serve", llog.KV{"err": err})
			}
			llog.Debug("http server closed")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	llog.Info("shutting down")
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(ctx, shutDownGracePeriod)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err = srv.Shutdown(ctx); err != nil {
		llog.Fatal("http server shutdown failed", llog.KV{"err": err})
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	os.Exit(0)
}

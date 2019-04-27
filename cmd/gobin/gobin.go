package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gobin"
	"github.com/levenlabs/go-llog"
)

var (
	shutDownGracePeriod = time.Second * 120
	textTmplsPath       = "./templates/textTemplates.tmpl"
	htmlTmplsPath       = "./templates/htmlTemplates.tmpl"
	domain              = "127.0.0.1:8081"
	staticDir           = "./static/"
)

func main() {

	ctx := context.Background()
	llog.SetLevelFromString("DEBUG")

	rand.Seed(time.Now().UTC().UnixNano())

	tmpls, err := gobin.NewTemplates(htmlTmplsPath, textTmplsPath, domain)
	if err != nil {
		llog.Fatal("failed to load templates", llog.ErrKV(err))
	}

	db, err := db.Connect(ctx, "host=127.0.0.1 port=26257 user=gobin dbname=gobin sslmode=disable")
	if err != nil {
		llog.Fatal("failed to connect to db", llog.KV{"err": err})
	}

	r := mux.NewRouter()
	routeToDir(r, "/browserconfig.xml", staticDir)
	routeToDir(r, "/robots.txt", staticDir)
	routeToDir(r, "/sitemap.xml", staticDir)

	r.Handle("/", gobin.GetRootHandler(db, tmpls)).Methods("GET")
	r.Handle("/", gobin.PostGobHandler(db, tmpls)).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	r.Handle("/{id:[a-zA-Z0-9]}", gobin.GetGobHandler(db)).Methods("GET")
	r.Handle("/expire/{secret}", gobin.GetExpireHandler(db, tmpls)).Methods("GET")
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
		Addr: "127.0.0.1:8081",
		// Need to figure out how high to set this
		WriteTimeout: time.Second * 120,
		ReadTimeout:  time.Second * 120,
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
	llog.Flush()
	os.Exit(0)
}

func routeToDir(r *mux.Router, path string, dir string) {
	r.PathPrefix(path).Handler(http.FileServer(http.Dir(dir)))
}

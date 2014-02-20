package main

import (
	"github.com/bmizerany/pat"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/conf"
	"github.com/kinghrothgar/goblin/handler"
	"github.com/kinghrothgar/goblin/storage/store"
	"net/http"
	"os"
	"os/signal"
)

// To be set at build
var buildCommit string
var buildDate string

func main() {
	if conf.ShowVers {
		println("Commit: " + buildCommit)
		println("Date:   " + buildDate)
		os.Exit(0)
	}

	conf.Parse()
	gslog.SetMinimumLevel(conf.LogLevel)
	//gslog.SetLogFile(conf.LogFile)

	store.Initialize(conf.StoreType, "", conf.UIDLen)

	gslog.Info("Goblin started [build commit: %s, build date: %s]", buildCommit, buildDate)

	// Setup route handlers
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(handler.GetRoot))
	// Could be horde or uid
	mux.Get("/:uid", http.HandlerFunc(handler.GetGob))
	mux.Get("/h/:horde", http.HandlerFunc(handler.GetHorde))
	mux.Post("/", http.HandlerFunc(handler.PostGob))
	mux.Post("/:horde", http.HandlerFunc(handler.PostHordeGob))
	// idkey is and string that contains the id and possibly an api key
	// TODO: will it contain gets?
	mux.Del("/:uidkey", http.HandlerFunc(handler.DelGob))

	http.Handle("/", mux)

	gslog.Info("Listening...")
	http.ListenAndServe(":3000", nil)
	if err := http.ListenAndServe(":3000", nil); err != nil {
		gslog.Error("ListenAndServe: %s", err)
		gslog.Fatal("Failed to start server, exiting...")
	}

	// Set up listening for os signals
	sigCh := make(chan os.Signal, 5)
	// TODO: What signals for Windows if any?
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	<-sigCh
	println("testing")
	gslog.Info("Syscall recieved, shutting down...")
	gslog.Flush()
	os.Exit(0)
}

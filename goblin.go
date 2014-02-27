package main

import (
	"github.com/bmizerany/pat"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/conf"
	"github.com/kinghrothgar/goblin/handler"
	"github.com/kinghrothgar/goblin/storage/store"
	"github.com/kinghrothgar/goblin/templ"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// To be set at build
var buildCommit string
var buildDate string

func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func listenAndServer(addr string, c chan error) {
	err := http.ListenAndServe(addr, nil)
	c <- err
}

func main() {
	if conf.ShowVers {
		println("Commit: " + buildCommit)
		println("Date:   " + buildDate)
		os.Exit(0)
	}

	gslog.Info("Goblin started [build commit: %s, build date: %s]", buildCommit, buildDate)

	conf.Parse()
	gslog.SetMinimumLevel(conf.LogLevel)
	//gslog.SetLogFile(conf.LogFile)

	if err := store.Initialize(conf.StoreType, "", conf.UIDLen); err != nil {
		gslog.Fatal("failed to initialize storage with error: %s", err.Error())
	}
	if err := templ.Initialize(conf.HTMLTemplatesPath, conf.TextTemplatesPath, conf.Domain); err != nil {
		gslog.Fatal("failed to initialize templates with error: %s", err.Error())
	}

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

	// Mandatory root-based resources
	serveSingle("/sitemap.xml", filepath.Join(conf.StaticPath, "sitemap.xml"))
	serveSingle("/favicon.ico", filepath.Join(conf.StaticPath, "favicon.ico"))
	serveSingle("/robots.txt", filepath.Join(conf.StaticPath, "robots.txt"))

	// Normal static resources
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(conf.StaticPath))))

	gslog.Info("Listening on " + conf.Port + " ...")
	c := make(chan error)
	go listenAndServer(":"+conf.Port, c)

	// Set up listening for os signals
	shutdownCh := make(chan os.Signal, 5)
	// TODO: What signals for Windows if any?
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGKILL)
	// Set up listening for os signals
	reloadCh := make(chan os.Signal, 5)
	signal.Notify(reloadCh, syscall.SIGUSR2)
	for {
		select {
		case <-reloadCh:
			if err := templ.Reload(conf.HTMLTemplatesPath, conf.TextTemplatesPath, conf.Domain); err != nil {
				gslog.Error("failed to reload with error: %s", err.Error())
			}
		case <-shutdownCh:
			gslog.Info("Syscall recieved, shutting down...")
			gslog.Flush()
			os.Exit(0)
		case err := <-c:
			gslog.Error("ListenAndServe: %s", err)
			gslog.Fatal("Failed to start server, exiting...")
		}
	}
}

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
	//if conf.ShowVers {
	//	println("Commit: " + buildCommit)
	//	println("Date:   " + buildDate)
	//	os.Exit(0)
	//}

	gslog.Info("Goblin started [build commit: %s, build date: %s]", buildCommit, buildDate)

	if err := conf.Parse(); err != nil {
		gslog.Fatal("MAIN: failed to parse conf with error: %s", err.Error())
	}

	gslog.SetMinimumLevel(conf.GetStr("loglevel"))
	if logFile := conf.GetStr("logfile"); logFile != "" {
		gslog.SetLogFile(logFile)
	}

	storeType, storeConf := conf.GetStr("storetype"), conf.GetStr("storeconf")
	uidLen, delUIDLen := conf.GetInt("uidlength"), conf.GetInt("deluidlength")
	handler.Initialize(uidLen, delUIDLen)
	if err := store.Initialize(storeType, storeConf, uidLen, delUIDLen); err != nil {
		gslog.Fatal("MAIN: failed to initialize storage with error: %s", err.Error())
	}
	htmlTemps, textTemps := conf.GetStr("htmltemplates"), conf.GetStr("texttemplates")
	domain := conf.GetStr("domain")
	if err := templ.Initialize(htmlTemps, textTemps, domain); err != nil {
		gslog.Fatal("MAIN: failed to initialize templates with error: %s", err.Error())
	}

	// Setup route handlers
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(handler.GetRoot))
	mux.Get("/:uid", http.HandlerFunc(handler.GetGob))
	mux.Get("/delete/:delUID", http.HandlerFunc(handler.DelGob))
	mux.Get("/horde/:horde", http.HandlerFunc(handler.GetHorde))
	mux.Post("/", http.HandlerFunc(handler.PostGob))
	// TODO: Should I post to /horde/:horde
	mux.Post("/:horde", http.HandlerFunc(handler.PostHordeGob))

	http.Handle("/", mux)

	// Mandatory root-based resources
	staticPath := conf.GetStr("staticpath")
	serveSingle("/sitemap.xml", filepath.Join(staticPath, "sitemap.xml"))
	serveSingle("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))
	serveSingle("/robots.txt", filepath.Join(staticPath, "robots.txt"))

	// Normal static resources
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(staticPath))))

	listenOn := conf.GetStr("listen")
	gslog.Info("MAIN: Listening on %s...", listenOn)
	c := make(chan error)
	go listenAndServer(listenOn, c)

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
			gslog.Info("MAIN: reloading")
			if err := conf.Parse(); err != nil {
				gslog.Error("MAIN: failed to parse conf with error: %s", err.Error())
				break
			}

			gslog.SetMinimumLevel(conf.GetStr("loglevel"))
			if logFile := conf.GetStr("logfile"); logFile != "" {
				gslog.SetLogFile(logFile)
			}

			storeConf = conf.GetStr("storeconf")
			uidLen, delUIDLen = conf.GetInt("uidlength"), conf.GetInt("deluidlength")
			handler.Initialize(uidLen, delUIDLen)
			store.Configure(storeConf, uidLen, delUIDLen)

			htmlTemps, textTemps = conf.GetStr("htmltemplates"), conf.GetStr("texttemplates")
			domain = conf.GetStr("domain")
			if err := templ.Reload(htmlTemps, textTemps, domain); err != nil {
				gslog.Error("MAIN: failed to reload templates with error: %s", err.Error())
			}
		case <-shutdownCh:
			gslog.Info("MAIN: Syscall recieved, shutting down...")
			gslog.Flush()
			os.Exit(0)
		case err := <-c:
			gslog.Error("MAIN: ListenAndServe: %s", err)
			gslog.Fatal("MAIN: Failed to start server, exiting...")
		}
	}
}

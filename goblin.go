package main

import (
	"flag"
	"github.com/bmizerany/pat"
	"github.com/grooveshark/golib/gslog"
	"log"
	"net/http"
)

// To be set at build
var buildCommit string
var buildDate string

type errorHistory map[string][]error

// Variables for flags
var (
	//	generalPath string
	//	logsPath    string
	logLevel string
	//	logFile     string
	showVers bool
)

func init() {
	// Find location of binary
	//exePath, _ := os.Readlink("/proc/self/exe")
	//path, _ := filepath.Split(exePath)
	//flag.StringVar(&generalPath, "gconf", filepath.Join(path, "general.yaml"), "path to general config file")
	//flag.StringVar(&logsPath, "lconf", filepath.Join(path, "logs.yaml"), "path to logs config file")
	//flag.StringVar(&logLevel, "loglevel", "", "level logging (DEBUG, INFO, WARN, ERROR, FATAL)")
	//flag.StringVar(&logFile, "logfile", "", "path to log file")
	flag.BoolVar(&showVers, "v", false, "show version/build information")
	flag.Parse()
	logLevel = "DEBUG"
	switch logLevel {
	case "", "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
		break
	default:
		gslog.Fatal("Invalid loglevel flag argument")
	}
}

func main() {
	mux := pat.New()
	mux.Get("/user/:name/profile", http.HandlerFunc(profile))

	http.Handle("/", mux)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func profile(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get(":name")
	w.Write([]byte("Hello " + name))
}

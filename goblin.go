package main

import (
	"bitbucket.org/kardianos/osext"
	"crypto/rand"
	"flag"
	"github.com/bmizerany/pat"
	"github.com/grooveshark/golib/gslog"
	//"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	//"fmt"
	"time"
)

const (
	ALPHA = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Gob struct {
	Id      string
	Text    string
	Created time.Time
}

type Gobs map[string]*Gob

// To be set at build
var buildCommit string
var buildDate string

// Variables for flags
var (
	//	generalPath string
	//	logsPath    string
	logLevel string
	logFile  string
	showVers bool
	gobs     = make(Gobs)
)

func init() {
	// Find location of binary
	//flag.StringVar(&generalPath, "gconf", filepath.Join(path, "general.yaml"), "path to general config file")
	//flag.StringVar(&logsPath, "lconf", filepath.Join(path, "logs.yaml"), "path to logs config file")
	flag.StringVar(&logLevel, "loglevel", "", "level logging (DEBUG, INFO, WARN, ERROR, FATAL)")
	flag.StringVar(&logFile, "logfile", "", "path to log file")
	flag.BoolVar(&showVers, "V", false, "show version/build information")
	flag.Parse()
	logLevel = strings.ToUpper(logLevel)
	switch logLevel {
	case "", "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
		break
	default:
		gslog.Fatal("Invalid loglevel flag argument")
	}
}

func main() {
	if showVers {
		println("Commit: " + buildCommit)
		println("Date:   " + buildDate)
		os.Exit(0)
	}

	if logLevel == "" {
		gslog.SetMinimumLevel("DEBUG")
	}

	if logFile == "" {
		exeFolder, _ := osext.ExecutableFolder()
		println("setting log file to " + filepath.Join(exeFolder, "goblin.log"))
		//gslog.SetLogFile(filepath.Join(exeFolder, "goblin.log"))
	}

	gslog.Info("Goblin started [build commit: %s, build date: %s]", buildCommit, buildDate)

	// Setup route handlers
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(index))
	mux.Post("/", http.HandlerFunc(post))
	mux.Get("/:id", http.HandlerFunc(getGob))

	http.Handle("/", mux)

	gslog.Info("Listening...")
	http.ListenAndServe(":3000", nil)

	// Set up listening for os signals
	sigCh := make(chan os.Signal, 5)
	// TODO: What signals for Windows if any?
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-sigCh:
			gslog.Info("Syscall recieved, shutting down...")
			gslog.Flush()
			os.Exit(0)
		}
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	//params := r.URL.Query()
	//name := params.Get(":name")
	w.Write([]byte("Hello"))
}

func post(w http.ResponseWriter, r *http.Request) {
	//parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//get a ref to the parsed multipart form
	m := r.MultipartForm
	if _, ok := m.Value["gob"]; !ok {
		gslog.Debug("post missing gob key")
		http.Error(w, "post missing gob key", http.StatusBadRequest)
		return
	}
	text := m.Value["gob"][0]
	id := getRandUniqStr()
	gobs[id] = &Gob{Id: id, Text: text, Created: time.Now()}

	w.Write([]byte("http://127.0.0.1:3000/" + id))
}

func getGob(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get(":id")
	if _, ok := gobs[id]; !ok {
		gslog.Debug("id does not exist")
		http.Error(w, "id does not exist", http.StatusNotFound)
		return
	}
	w.Write([]byte(gobs[id].Text))
}

func getRandUniqStr() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	str := string(bytes)
	if _, ok := gobs[str]; ok {
		str = getRandUniqStr()
	}
	return str
}

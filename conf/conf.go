package conf

import (
	"bitbucket.org/kardianos/osext"
	"flag"
	//"github.com/grooveshark/golib/gslog"
	"errors"
	"path/filepath"
	"strings"
)

// Variables for flags
var (
	ConfPath  string
	LogLevel  string
	LogFile   string
	ShowVers  bool
	ExeFolder string
	UIDLen    int
	StoreType string
	Domain    string
	Favicon   string
	Port      string
)

func init() {
	// Find location of binary
	ExeFolder, _ = osext.ExecutableFolder()
	flag.StringVar(&ConfPath, "config", filepath.Join(ExeFolder, "goblin.yaml"), "path to general config file")
	flag.StringVar(&LogLevel, "loglevel", "", "level logging (DEBUG, INFO, WARN, ERROR, FATAL)")
	flag.StringVar(&LogFile, "logfile", "", "path to log file")
	flag.BoolVar(&ShowVers, "V", false, "show version/build information")
	flag.Parse()
	LogLevel = strings.ToUpper(LogLevel)
}

func Parse() error {
	if LogLevel == "" {
		LogLevel = "DEBUG"
	}

	if LogFile == "" {
		LogFile = filepath.Join(ExeFolder, "goblin.log")
	}

	UIDLen = 4
	StoreType = "REDIS"
	Domain = "gobin.io"
	Port = "6667"
	return nil
}

func Validate() error {
	switch LogLevel {
	case "", "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
		break
	default:
		return errors.New("Invalid loglevel flag argument")
	}
	return nil
}

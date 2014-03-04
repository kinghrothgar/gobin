package conf

import (
	"bitbucket.org/kardianos/osext"
	//"github.com/grooveshark/golib/gslog"
	"github.com/mediocregopher/flagconfig"
	"path/filepath"
	"sync"
)

// Variables for flags
var (
	fc     *flagconfig.FlagConfig
	fcLock = sync.RWMutex{}
)

func Parse() error {
	exeFolder, _ := osext.ExecutableFolder()

	f := flagconfig.New("goblin")
	f.StrParam("loglevel", "logging level (DEBUG, INFO, WARN, ERROR, FATAL)", "DEBUG")
	f.StrParam("logfile", "path to log file", "")
	f.StrParam("htmltemplates", "path to html templates file", filepath.Join(exeFolder, "templates/htmlTemplates.tmpl"))
	f.StrParam("texttemplates", "path to text templates file", filepath.Join(exeFolder, "templates/textTemplates.tmpl"))
	f.StrParam("staticpath", "path to static files folder", filepath.Join(exeFolder, "static"))
	f.IntParam("uidlength", "length of gob uid string", 4)
	f.IntParam("deluidlength", "length of the delete uid string", 15)
	f.RequiredStrParam("storetype", "the data store to use")
	f.RequiredStrParam("storeconf", "a string of the form 'IP:PORT' to configure the data store")
	f.RequiredStrParam("domain", "the domain to use to for links")
	f.RequiredStrParam("listen", "a string of the form 'IP:PORT' which program will listen on")
	f.FlagParam("V", "show version/build information", false)

	if err := f.Parse(); err != nil {
		return err
	}
	fcLock.Lock()
	defer fcLock.Unlock()
	fc = f
	//UIDLen = 4
	//StoreType = "REDIS"
	//Domain = "gobin.io"
	//Port = "6667"
	return nil
}

func GetStr(k string) string {
	fcLock.RLock()
	defer fcLock.RUnlock()
	return fc.GetStr(k)
}

func GetStrs(k string) []string {
	fcLock.RLock()
	defer fcLock.RUnlock()
	return fc.GetStrs(k)
}

func GetInt(k string) int {
	fcLock.RLock()
	defer fcLock.RUnlock()
	return fc.GetInt(k)
}

//func Validate() error {
//	switch LogLevel {
//	case "", "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
//		break
//	default:
//		return errors.New("Invalid loglevel flag argument")
//	}
//	return nil
//}

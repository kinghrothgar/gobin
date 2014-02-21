package handler

import (
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/conf"
	"github.com/kinghrothgar/goblin/storage/store"
	"net"
	"net/http"
	"regexp"
)

var (
	alphaReg = regexp.MustCompile("^[A-Za-z]+$")
)

func getGobData(w http.ResponseWriter, r *http.Request) []byte {
	//parse the multipart form in the request
	err := r.ParseMultipartForm(10485760)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	//get a ref to the parsed multipart form
	m := r.MultipartForm
	str := m.Value["gob"][0]
	return []byte(str)
}

func validateUID(w http.ResponseWriter, uid string) error {
	// This is so someone can't access a horde goblin
	// by just puting the 'horde#uid' instead of 'horde/uid'
	// and prevents a lookup if it's obviously crap
	if len(uid) > conf.UIDLen || !alphaReg.MatchString(uid) {
		err := errors.New("invalid uid")
		gslog.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}

func validateHordeName(w http.ResponseWriter, hordeName string) {
	// TODO: horde max length?
	if len(hordeName) > 50 || !alphaReg.MatchString(hordeName) {
		gslog.Debug("invalid horde name")
		http.Error(w, "invalid horde name", http.StatusBadRequest)
	}
	return
}

func formURL(uid string) string {
	return "http://" + conf.Domain + "/" + uid + "\n"
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	page := ""
	page += "Welcome to GoBin, command line pastebin.\n"
	page += "Backend using goblin written in go and redis\n\n"
	page += "<command> | curl -F 'gob=<-' gobin.io\n"
	page += "Or, to paste to a horde:\n"
	page += "<command> | curl -F 'gob=<-' gobin.io/<horde>\n"
	page += "Going to gobin.io/h/<horde> will list everything that has been pasted to it\n"
	w.Write([]byte(page))
}

func GetGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetGob called")
	params := r.URL.Query()
	uid := params.Get(":uid")
	if err := validateUID(w, uid); err != nil {
		gslog.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, _, err := store.GetGob(uid)
	if err != nil {
		gslog.Debug("failed to get gob with error: " + err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	gslog.Debug("GetGob writing data")
	w.Write(data)
}

func GetHorde(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetHorde called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	uidTimeList, err := store.GetHorde(hordeName)
	if err != nil {
		gslog.Debug("failed to get horde with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := ""
	for _, uidTimePair := range uidTimeList {
		url := formURL(uidTimePair.UID)
		page += url + "    " + uidTimePair.Time.String() + "\n"
	}
	w.Write([]byte(page))
}

func PostGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("PostGob called")
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	gslog.Debug("uid: %s, host: %s, ip: %s", uid, host, ip)
	if err := store.PutGob(uid, gobData, ip); err != nil {
		gslog.Debug("put gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := formURL(uid)
	w.Write([]byte(url))
}

func PostHordeGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("PostHordeGob called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	validateHordeName(w, hordeName)
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	gslog.Debug("uid: %s, ip: %s", uid, ip)
	if err := store.PutHordeGob(uid, hordeName, gobData, ip); err != nil {
		gslog.Debug("put horde gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	url := formURL(uid)
	w.Write([]byte(url))
}

func DelGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("DelGob called")
}

func DelHordeGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("DelHordeGob called")
}

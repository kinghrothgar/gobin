package handler

import (
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage/store"
	"net"
	"net/http"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func GetUID(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uid := params.Get(":uid")
	data, _, err := store.GetGob(uid)
	if err != nil {
		gslog.Debug("id does not exist")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(data)
}

func GetHorde(w http.ResponseWriter, r *http.Request) {
}

func GetHordeUID(w http.ResponseWriter, r *http.Request) {
}

func PostRoot(w http.ResponseWriter, r *http.Request) {
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
	str := m.Value["gob"][0]
	uid := store.GetRandUID()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	gslog.Debug("uid: %s, host: %s, ip: %s", uid, host, ip)
	if err = store.PutGob(uid, []byte(str), ip); err != nil {
		gslog.Debug("put gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write([]byte("http://127.0.0.1:3000/" + uid))
}

func PostHorde(w http.ResponseWriter, r *http.Request) {
}

func DelUID(w http.ResponseWriter, r *http.Request) {
}

func DelHordeUID(w http.ResponseWriter, r *http.Request) {
}

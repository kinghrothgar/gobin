package gobin

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gob"
	"github.com/levenlabs/go-llog"
)

// TODO investigate whether curl loads file into memory when using @ or @-
func PostGobHandler(db *db.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, gobHeader, err := r.FormFile("gob")
		if err != nil {
			llog.Debug("failed to get form file gob", llog.KV{"err": err})
			returnHTTPBadRequest(w, "request must have form file 'gob'")
			return
		}
		// TODO return error if gobHeader.Size too big
		llog.Debug("got file upload", llog.KV{"filename": gobHeader.Filename, "size": gobHeader.Size})
		gobFile, err := gobHeader.Open()
		if err != nil {
			llog.Fatal("failed to open file gob", llog.KV{"err": err})
		}
		defer gobFile.Close()
		encryptKey := r.URL.Query().Get("encrypt")
		gob := gob.NewGob(r.Context(), db)
		meta, err := gob.Upload(gobFile, encryptKey)
		if err != nil {
			llog.Error("failed to upload gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to upload gob")
			return
		}
		response := fmt.Sprintf("%s\ndelete/%s\n", meta.ID, meta.AuthKey)
		w.Write([]byte(response))
		llog.Debug("uploaded gob", llog.KV{"id": meta.ID})
	})
}

// TODO investigate whether curl loads file into memory when using @ or @-
func GetGobHandler(db *db.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			llog.Error("failed to get id in GetGobHandler")
			returnHTTPBadRequest(w, "request must have gob id in url")
			return
		}
		// TODO validate id
		encryptKey := r.URL.Query().Get("encrypt")
		gob := gob.NewGob(r.Context(), db)
		meta, err := gob.Download(w, id, encryptKey)
		// TODO figure out if it was user error
		if err != nil {
			llog.Error("failed to download gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to download gob")
			return
		}
		llog.Debug("downloaded gob", llog.KV{"id": meta.ID})
	})
}

func returnHTTPInternalError(w http.ResponseWriter, message string) {
	http.Error(w, "Error: "+message, http.StatusInternalServerError)
}

func returnHTTPBadRequest(w http.ResponseWriter, message string) {
	http.Error(w, "Error: "+message, http.StatusBadRequest)
}

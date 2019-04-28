package gobin

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kinghrothgar/gobin/pkg/db"
	"github.com/kinghrothgar/gobin/pkg/gob"
	"github.com/levenlabs/go-llog"
)

var (
	// TODO pass around valid characters config
	alphaReg            = regexp.MustCompile("^[A-Za-z]+$")
	alphaNumericReg     = regexp.MustCompile("^[A-Za-z0-9]+$")
	browserUserAgentReg = regexp.MustCompile("Mozilla")
	textContentTypeReg  = regexp.MustCompile("^text/")
)

func GetRootHandler(db *db.DB, tmpls *Templates) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		llog.Debug("GetRootHandler called with header", llog.KV{"header": r.Header, "host": r.Host, "requestURI": r.RequestURI, "remoteAddr": r.RemoteAddr})
		pageType := getPageType(r)
		pageBytes, err := tmpls.GetHomePage(pageType)
		if err != nil {
			llog.Error("failed to get home", llog.ErrKV(err))
			returnHTTPInternalError(w, "failed to get home")
			return
		}
		w.Write(pageBytes)
	})
}

// TODO investigate whether curl loads file into memory when using @ or @-
func PostGobHandler(db *db.DB, tmpls *Templates) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, gobHeader, err := r.FormFile("g")
		if err != nil {
			llog.Debug("failed to get form file gob", llog.KV{"err": err})
			returnHTTPBadRequest(w, "request must have form file 'gob'")
			return
		}
		filename := gobHeader.Filename
		if fn := r.FormValue("f"); fn != "" {
			filename = fn
		}
		// TODO return error if gobHeader.Size too big
		llog.Debug("got file upload", llog.KV{"filename": filename, "size": gobHeader.Size})
		gobFile, err := gobHeader.Open()
		if err != nil {
			llog.Fatal("failed to open file gob", llog.KV{"err": err})
		}
		defer gobFile.Close()
		encryptKey := r.URL.Query().Get("encrypt")
		gob := gob.NewGob(r.Context(), db)
		meta, err := gob.Upload(gobFile, encryptKey, filename)
		if err != nil {
			llog.Error("failed to upload gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to upload gob")
			return
		}
		pageType := getPageType(r)
		pageBytes, err := tmpls.GetURLPage(getScheme(r), pageType, meta.ID, meta.Secret)
		// TODO should delete gob if we can't tell users the id
		if err != nil {
			llog.Error("failed to upload gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to upload gob")
			return
		}
		w.Write(pageBytes)
		llog.Debug("uploaded gob", llog.KV{"id": meta.ID})
	})
}

// TODO investigate whether curl loads file into memory when using @ or @-
// TODO validate gob id
func GetGobHandler(db *db.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			llog.Debug("failed to get id in GetGobHandler")
			returnHTTPBadRequest(w, "request must have gob id in url")
			return
		}
		// TODO validate id
		encryptKey := r.URL.Query().Get("encrypt")
		gob := gob.NewGob(r.Context(), db)
		meta, err := gob.GetMetadata(id)
		// TODO figure out if it was user error
		if err != nil {
			returnHTTPNotFound(w, id+" gob not found")
		}
		// TODO will cause download in browser
		//if meta.Filename.Valid {
		//	w.Header().Set("Content-Disposition", "attachment; filename="+meta.Filename.String)
		//}
		w.Header().Set("Content-Type", meta.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
		// TODO figure out if it was user error
		if err = gob.Download(w, meta, encryptKey); err != nil {
			llog.Error("failed to download gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to download gob")
			return
		}
		llog.Debug("downloaded gob", llog.KV{"id": meta.ID})
	})
}

func GetExpireHandler(db *db.DB, tmpls *Templates) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		secret, ok := vars["secret"]
		if !ok {
			llog.Error("failed to get secret in GetExpireHandler")
			returnHTTPBadRequest(w, "expire request must have gob secret in url")
			return
		}
		// TODO validate id
		gob := gob.NewGob(r.Context(), db)
		meta, err := gob.Expire(secret)
		// TODO figure out if it was user error
		if err != nil {
			llog.Warn("failed to expire gob", llog.KV{"err": err})
			returnHTTPInternalError(w, "failed to expire gob")
			return
		}

		pageType := getPageType(r)
		pageBytes, err := tmpls.GetMessPage(pageType, "successfully deleted "+meta.ID)
		w.Write(pageBytes)
		llog.Debug("expired gob", llog.KV{"id": meta.ID})
	})
}

func returnHTTPError(w http.ResponseWriter, message string, status int) {
	http.Error(w, "Error: "+message, status)
}

func returnHTTPInternalError(w http.ResponseWriter, message string) {
	http.Error(w, "Error: "+message, http.StatusInternalServerError)
}

func returnHTTPNotFound(w http.ResponseWriter, message string) {
	http.Error(w, "Error: "+message, http.StatusNotFound)
}

func returnHTTPBadRequest(w http.ResponseWriter, message string) {
	http.Error(w, "Error: "+message, http.StatusBadRequest)
}

func getScheme(r *http.Request) (scheme string) {
	hdr := r.Header
	if scheme = hdr.Get("X-Real-Scheme"); scheme == "" {
		scheme = "http"
	}
	return
}

func getPageType(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	params := r.URL.Query()
	_, cli := params["cli"]
	// If cli param present or Mozilla not found in the user agent, use plain text
	if cli || !browserUserAgentReg.MatchString(userAgent) {
		return "TEXT"
	}
	return "HTML"
}

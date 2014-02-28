package handler

import (
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage/store"
	"github.com/kinghrothgar/goblin/templ"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var (
	alphaReg            = regexp.MustCompile("^[A-Za-z]+$")
	browserUserAgentReg = regexp.MustCompile("Mozilla")
	textContentTypeReg  = regexp.MustCompile("^text/")
	uidLen int
)

func getGobData(w http.ResponseWriter, r *http.Request) []byte {
	//parse the multipart form in the request
	err := r.ParseMultipartForm(11534336) // 11 MB
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	//get a ref to the parsed multipart form
	m := r.MultipartForm
	str := m.Value["gob"][0]
	return []byte(str)
}

func validUID(uid string) bool {
	// This is so someone can't access a horde goblin
	// by just puting the 'horde#uid' instead of 'horde/uid'
	// and prevents a lookup if it's obviously crap
	if len(uid) > uidLen || !alphaReg.MatchString(uid) {
		return false
	}
	return true
}

func validHordeName(hordeName string) bool {
	// TODO: horde max length?
	if len(hordeName) > 50 || !alphaReg.MatchString(hordeName) {
		return false
	}
	return true
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	host, _, _ := net.SplitHostPort(s)
	return net.ParseIP(host).String()
}

func returnHTTPError(w http.ResponseWriter, funcName string, errMessage string, status int) {
	gslog.Debug("HANDLER: %s returned error to user: %s", funcName, errMessage)
	http.Error(w, errMessage, http.StatusInternalServerError)
}

func getIpAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
		        parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return net.ParseIP(hdrRealIp).String()
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

func getLanguage(r *http.Request) string {
	params := r.URL.Query()
	// Firgure out language parameter
	lang := ""
	for key, _ := range params {
		// Skip the pat params that start with :
		if key[0] == ':' {
			continue
		}
		lang = strings.ToLower(key)
	}
	// Deal with aliases
	switch lang {
	case "javascript", "js":
		lang = "javascript"
	    break
	case "coffeescript", "coffee":
		lang = "coffeescript"
	    break
	case "bash", "sh":
		lang = "bash"
	    break
	case "python", "py":
		lang = "python"
	    break
	case "groovy", "gvy", "gy", "gsh":
		lang = "groovy"
	    break
	case "ruby", "rb":
		lang = "ruby"
	    break
	case "markdown", "md":
		lang = "markdown"
		break
	case "c":
	case "clike":
	case "cpp":
	case "csharp":
	case "css":
	case "css.selector":
	case "gherkin":
	case "go":
	case "http":
	case "java":
	case "markup":
	case "php":
	case "scss":
	case "sql":
	    break
	default:
		lang = ""
	}
	return lang
}

// Must be called before any other functions are called
// TODO: should I just called it SetUIDLen ?
func Initialize(uidLength int) {
	uidLen = uidLength
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetRoot called")
	params := r.URL.Query()
	gslog.Debug("query params: %+v", params)
	pageType := getPageType(r)
	pageBytes, err := templ.GetHomePage(pageType)
	if err != nil {
		gslog.Error("HANDLER: failed to get home with error: %s", err.Error())
		returnHTTPError(w, "GetRoot", "failed to get home", http.StatusInternalServerError)
		return
	}
	w.Write(pageBytes)
}

func GetGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetGob called")
	params := r.URL.Query()
	uid := params.Get(":uid")
	if !validUID(uid) {
		returnHTTPError(w, "GetGob", uid + " not found", http.StatusNotFound)
		return
	}
	data, _, err := store.GetGob(uid)
	if err != nil {
		gslog.Error("HANDLER: failed to get gob with error: " + err.Error())
		returnHTTPError(w, "GetGob", "failed to get gob", http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		returnHTTPError(w, "GetGob", uid + " not found", http.StatusNotFound)
		return
	}
	// Firgure out language parameter
	lang := getLanguage(r)

	if lang == "" {
		gslog.Debug("HANDLER: GetGob writing data")
		w.Write(data)
		return
	}
	contentType := http.DetectContentType(data)
	// If data is a valid content type for syntax highlighting
	if textContentTypeReg.MatchString(contentType) {
		data, err = templ.GetGobPage(lang, data)
		if err != nil {
			gslog.Error("HANDLER: failed to get gob page with error: %s", err.Error())
			returnHTTPError(w, "GetGob", "failed to get gob", http.StatusInternalServerError)
			return
		}
	}

	w.Write(data)
}

func GetHorde(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("HANDLER: GetHorde called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	if !validHordeName(hordeName) {
		returnHTTPError(w, "GetHorde", hordeName + " not found", http.StatusNotFound)
		return
	}
	horde, err := store.GetHorde(hordeName)
	if err != nil {
		gslog.Error("HANDLER: failed to get horde with error: %s", err.Error())
		returnHTTPError(w, "GetHorde", "failed to get horde", http.StatusInternalServerError)
		return
	}
	if len(horde) == 0 {
		returnHTTPError(w, "GetHorde", hordeName + " not found", http.StatusNotFound)
		return
	}
	pageType := getPageType(r)
	pageBytes, err := templ.GetHordePage(pageType, hordeName, horde)
	if err != nil {
		gslog.Debug("HANDLER: failed to get horde with error: %s", err.Error())
		returnHTTPError(w, "GetHorde", "failed to get horde", http.StatusInternalServerError)
		return
	}
	w.Write(pageBytes)
}

func PostGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("HANDLER: PostGob called with header: %+v, host: %s, requestURI: %s", r.Header, r.Host, r.RequestURI)
	gobData := getGobData(w, r)
	if len(gobData) == 0 {
		returnHTTPError(w, "PostGob", "gob empty", http.StatusBadRequest)
		return
	}

	uid := store.GetNewUID()
	ip := getIpAddress(r)
	gslog.Debug("HANDLER: uid: %s, ip: %s", uid, ip)
	if err := store.PutGob(uid, gobData, ip); err != nil {
		gslog.Error("HANDLER: put gob failed with error: %s", err.Error())
		returnHTTPError(w, "PostGob", "failed to save gob", http.StatusInternalServerError)
		return
	}

	url := templ.BuildURL(uid)
	w.Write([]byte(url))
}

func PostHordeGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("PostHordeGob called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	if !validHordeName(hordeName) {
		returnHTTPError(w, "PostHordeGob", hordeName + " not found", http.StatusNotFound)
		return
	}
	gobData := getGobData(w, r)
	if len(gobData) == 0 {
		returnHTTPError(w, "PostHordeGob", "gob empty", http.StatusBadRequest)
		return
	}
	uid := store.GetNewUID()
	ip := getIpAddress(r)
	gslog.Debug("uid: %s, ip: %s", uid, ip)
	if err := store.PutHordeGob(uid, hordeName, gobData, ip); err != nil {
		gslog.Error("HANDLER: put horde gob failed with error: %s", err.Error())
		returnHTTPError(w, "PostHordeGob", "failed to save gob", http.StatusInternalServerError)
	}

	url := templ.BuildURL(uid)
	w.Write([]byte(url))
}

func DelGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("DelGob called")
}

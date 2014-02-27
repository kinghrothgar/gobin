package handler

import (
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/conf"
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

func validateUID(uid string) error {
	// This is so someone can't access a horde goblin
	// by just puting the 'horde#uid' instead of 'horde/uid'
	// and prevents a lookup if it's obviously crap
	if len(uid) > conf.UIDLen || !alphaReg.MatchString(uid) {
		gslog.Debug("invalid uid")
		return errors.New("invalid uid")
	}
	return nil
}

func validateHordeName(hordeName string) error {
	// TODO: horde max length?
	if len(hordeName) > 50 || !alphaReg.MatchString(hordeName) {
		gslog.Debug("invalid horde name")
		return errors.New("invalid horde name")
	}
	return nil
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	host, _, _ := net.SplitHostPort(s)
	return net.ParseIP(host).String()
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

func GetRoot(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetRoot called")
	params := r.URL.Query()
	gslog.Debug("query params: %+v", params)
	pageType := getPageType(r)
	pageBytes, err := templ.GetHomePage(pageType)
	if err != nil {
		gslog.Debug("failed to write home with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(pageBytes)
}

func GetGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetGob called")
	params := r.URL.Query()
	uid := params.Get(":uid")
	if err := validateUID(uid); err != nil {
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
	// Firgure out language parameter
	lang := getLanguage(r)

	if lang == "" {
		gslog.Debug("GetGob writing data")
		w.Write(data)
		return
	}
	contentType := http.DetectContentType(data)
	// If data is a valid content type for syntax highlighting
	if textContentTypeReg.MatchString(contentType) {
		data, err = templ.GetGobPage(lang, data)
		if err != nil {
			gslog.Debug("failed to get gob page with error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Write(data)
}

func GetHorde(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("GetHorde called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	horde, err := store.GetHorde(hordeName)
	if err != nil {
		gslog.Debug("failed to get horde with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pageType := getPageType(r)
	pageBytes, err := templ.GetHordePage(pageType, hordeName, horde)
	if err != nil {
		gslog.Debug("failed to get horde with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(pageBytes)
}

func PostGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("PostGob called")
	gslog.Debug("Request has header: %+v", r.Header)
	gslog.Debug("Request has host: %s", r.Host)
	gslog.Debug("Request has requestURI: %s", r.RequestURI)
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	ip := getIpAddress(r)
	gslog.Debug("uid: %s, ip: %s", uid, ip)
	if err := store.PutGob(uid, gobData, ip); err != nil {
		gslog.Debug("put gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := templ.BuildURL(uid)
	w.Write([]byte(url))
}

func PostHordeGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("PostHordeGob called")
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	if err := validateHordeName(hordeName); err != nil {
		gslog.Debug(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	ip := getIpAddress(r)
	gslog.Debug("uid: %s, ip: %s", uid, ip)
	if err := store.PutHordeGob(uid, hordeName, gobData, ip); err != nil {
		gslog.Debug("put horde gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	url := templ.BuildURL(uid)
	w.Write([]byte(url))
}

func DelGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("DelGob called")
}

func DelHordeGob(w http.ResponseWriter, r *http.Request) {
	gslog.Debug("DelHordeGob called")
}

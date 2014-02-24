package templ

import (
	"github.com/kinghrothgar/goblin/storage"
	"html/template"
	"net/http"
)

type HordePage struct {
	Domain string
	Horde  storage.Horde
}

var (
	templates *template.Template
	domain    string
)

func Initialize(templatesPath string, confDomain string) error {
	var err error
	templates, err = template.ParseFiles(templatesPath)
	domain = confDomain
	return err
}

func WriteHordePage(w http.ResponseWriter, horde storage.Horde) error {
	p := &HordePage{Domain: domain, Horde: horde}
	return templates.ExecuteTemplate(w, "hordePage", p)
}

func BuildURL(uid string) string {
	return "http://" + domain + "/" + uid + "\n"
}

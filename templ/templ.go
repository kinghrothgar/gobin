package templ

import (
	"github.com/kinghrothgar/goblin/storage"
	htmlTemplate "html/template"
	textTemplate "text/template"
	"net/http"
	"errors"
)

type HordePage struct {
	Domain string
	Title  string
	Horde  storage.Horde
}

type HomePage struct {
	Domain string
	Title  string
	Horde  storage.Horde
}

var (
	htmlTemplates *htmlTemplate.Template
	textTemplates *textTemplate.Template
	domain    string
)

func executeTemplate(w http.ResponseWriter, contentType string, templateName string, data interface{}) error {
	switch contentType {
	case "HTML":
		return htmlTemplates.ExecuteTemplate(w, templateName, data)
	case "TEXT":
		return textTemplates.ExecuteTemplate(w, templateName, data)
	}
	return errors.New("invalid content type")
}

func unescaped (x string) interface{} {
	return htmlTemplate.HTML(x)
}

func Initialize(htmlTemplatesPath string, textTemplatesPath string, confDomain string) error {
	var err error
	htmlTemplates = htmlTemplate.New("homePage")
	htmlTemplates = htmlTemplates.Funcs(htmlTemplate.FuncMap{"unescaped": unescaped})
	htmlTemplates, err = htmlTemplate.ParseFiles(htmlTemplatesPath)
	if err != nil {
		return err
	}
	textTemplates, err = textTemplate.ParseFiles(textTemplatesPath)
	domain = confDomain
	return err
}

func WriteHordePage(w http.ResponseWriter, contentType string, hordeName string, horde storage.Horde) error {
	p := &HordePage{Domain: domain, Title: "horde: " + hordeName, Horde: horde}
	return executeTemplate(w, contentType, "hordePage", p)
}

func WriteHomePage(w http.ResponseWriter, contentType string) error {
	p := &HomePage{Domain: domain, Title: "gobin: a cli pastebin"}
	return executeTemplate(w, contentType, "homePage", p)
}

func BuildURL(uid string) string {
	return "http://" + domain + "/" + uid + "\n"
}

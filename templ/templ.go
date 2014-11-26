package templ

import (
	"bytes"
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/gobin/storage"
	"github.com/kinghrothgar/pygments"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

type Tabs struct {
	Home bool
	Form bool
	Top  bool
}

type HordePage struct {
	Domain string
	Scheme string
	Title  string
	Horde  storage.Horde
}

type HomePage struct {
	Domain string
	Title  string
	Tabs   *Tabs
}

type MessPage struct {
	Title   string
	Tabs    *Tabs
	Message string
}

type FormPage struct {
	Domain string
	Scheme string
	Title  string
	Tabs   *Tabs
}

type URLPage struct {
	Domain string
	Scheme string
	Title  string
	UID    string
	DelUID string
	Tabs   *Tabs
}

type GobPage struct {
	Title    string
	Language string
	Data     htmlTemplate.HTML
}

type MDPage struct {
	Title    string
	Language string
	Data     htmlTemplate.HTML
}

var (
	htmlTemplates  *htmlTemplate.Template
	textTemplates  *textTemplate.Template
	domain         string
	pygmentizePath string
)

func executeTemplate(contentType string, templateName string, data interface{}) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}
	switch contentType {
	case "HTML":
		err = htmlTemplates.ExecuteTemplate(buf, templateName, data)
		break
	case "TEXT":
		err = textTemplates.ExecuteTemplate(buf, templateName, data)
		break
	default:
		err = errors.New("invalid content type")
	}
	return buf.Bytes(), err
}

func unescaped(x string) interface{} {
	return htmlTemplate.HTML(x)
}

func Initialize(htmlTemplatesPath string, textTemplatesPath string, confDomain string, confPygmentizePath string) error {
	var err error
	htmlTemplates, err = htmlTemplate.ParseFiles(htmlTemplatesPath)
	if err != nil {
		return err
	}
	textTemplates, err = textTemplate.ParseFiles(textTemplatesPath)
	gslog.Debug("TEMPL: loaded htmlTemplates from %s", htmlTemplatesPath)
	gslog.Debug("TEMPL: loaded textTemplates from %s", textTemplatesPath)
	domain = confDomain
	pygmentizePath = confPygmentizePath
	return err
}

func Reload(htmlTemplatesPath string, textTemplatesPath string, confDomain string, confPygmentizePath string) error {
	if htmlTemplatesTemp, err := htmlTemplate.ParseFiles(htmlTemplatesPath); err != nil {
		return err
	} else {
		htmlTemplates = htmlTemplatesTemp
		gslog.Debug("htmlTemplates loaded")
	}
	if textTemplatesTemp, err := textTemplate.ParseFiles(textTemplatesPath); err != nil {
		return err
	} else {
		textTemplates = textTemplatesTemp
		gslog.Info("textTemplates loaded")
	}
	domain = confDomain
	pygmentizePath = confPygmentizePath
	return nil
}

func GetHordePage(scheme string, contentType string, hordeName string, horde storage.Horde) ([]byte, error) {
	p := &HordePage{Domain: domain, Scheme: scheme, Title: "horde: " + hordeName, Horde: horde}
	return executeTemplate(contentType, "hordePage", p)
}

func GetGobPage(language string, data []byte) ([]byte, error) {
	if language == "markdown" {
		p := &MDPage{
			Title:    "gob: " + language + " syntax highlighted",
			Language: language,
			Data:     htmlTemplate.HTML(string(data)),
		}
		return executeTemplate("HTML", "mdPage", p)
	}
	pygments.Binary(pygmentizePath)
	opts := pygments.Options{
		"linenos":  "table",
		"encoding": "utf-8",
	}
	code, err := pygments.Highlight(string(data), language, "html", opts)
	if err != nil {
		gslog.Error("Failed to highlight code: " + err.Error())
		// Syntax highlighting has failed, so just display the raw data
		return data, nil
	}
	p := &GobPage{Title: "gob: " + language + " syntax highlighted", Language: language, Data: htmlTemplate.HTML(code)}
	return executeTemplate("HTML", "gobPage", p)
}

func GetHomePage(contentType string) ([]byte, error) {
	t := &Tabs{Home: true}
	p := &HomePage{Domain: domain, Title: "gobin: a cli pastebin", Tabs: t}
	return executeTemplate(contentType, "homePage", p)
}

func GetFormPage(scheme string) ([]byte, error) {
	t := &Tabs{Form: true}
	p := &FormPage{Domain: domain, Scheme: scheme, Title: "gobin: a cli pastebin", Tabs: t}
	return executeTemplate("HTML", "formPage", p)
}

func GetMessPage(contentType string, message string) ([]byte, error) {
	t := &Tabs{}
	p := &MessPage{Title: "gobin: a cli pastebin", Tabs: t, Message: message}
	return executeTemplate(contentType, "messPage", p)
}

func GetURLPage(scheme, contentType, uid, delUID string) ([]byte, error) {
	t := &Tabs{Form: true}
	p := &URLPage{
		Domain: domain,
		Scheme: scheme,
		Title:  "gobin: a cli pastebin",
		UID:    uid,
		DelUID: delUID,
		Tabs:   t,
	}
	return executeTemplate(contentType, "urlPage", p)
}

// BuildURLs builds the urls given the scheme (http/https), uid and delUID
func BuildURLs(scheme, uid, delUID string) string {
	urls := scheme + "://" + domain + "/" + uid + "\n"
	urls += scheme + "://" + domain + "/delete/" + delUID + "\n"
	return urls
}

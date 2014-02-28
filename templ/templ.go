package templ

import (
	"bytes"
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

type HordePage struct {
	Domain string
	Title  string
	Horde  storage.Horde
}

type HomePage struct {
	Domain string
	Title  string
}

type GobPage struct {
	Title    string
	Language string
	Data     string
}

var (
	htmlTemplates *htmlTemplate.Template
	textTemplates *textTemplate.Template
	domain        string
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

func Initialize(htmlTemplatesPath string, textTemplatesPath string, confDomain string) error {
	var err error
	htmlTemplates, err = htmlTemplate.ParseFiles(htmlTemplatesPath)
	if err != nil {
		return err
	}
	textTemplates, err = textTemplate.ParseFiles(textTemplatesPath)
	gslog.Debug("TEMPL: loaded htmlTemplates from %s", htmlTemplatesPath)
	gslog.Debug("TEMPL: loaded textTemplates from %s", textTemplatesPath)
	domain = confDomain
	return err
}

func Reload(htmlTemplatesPath string, textTemplatesPath string, confDomain string) error {
	if htmlTemplatesTemp, err := htmlTemplate.ParseFiles(htmlTemplatesPath); err != nil {
		return err
	} else {
		htmlTemplates = htmlTemplatesTemp
		gslog.Info("htmlTemplates loaded")
	}
	if textTemplatesTemp, err := textTemplate.ParseFiles(textTemplatesPath); err != nil {
		return err
	} else {
		textTemplates = textTemplatesTemp
		gslog.Info("textTemplates loaded")
	}
	domain = confDomain
	return nil
}

func GetHordePage(contentType string, hordeName string, horde storage.Horde) ([]byte, error) {
	p := &HordePage{Domain: domain, Title: "horde: " + hordeName, Horde: horde}
	return executeTemplate(contentType, "hordePage", p)
}

func GetGobPage(language string, data []byte) ([]byte, error) {
	p := &GobPage{Title: "gob: " + language + " syntax highlighted", Language: language, Data: string(data)}
	if language == "markdown" {
		return executeTemplate("HTML", "mdPage", p)
	}
	return executeTemplate("HTML", "gobPage", p)
}

func GetHomePage(contentType string) ([]byte, error) {
	p := &HomePage{Domain: domain, Title: "gobin: a cli pastebin"}
	return executeTemplate(contentType, "homePage", p)
}

func BuildURL(uid string) string {
	return "http://" + domain + "/" + uid + "\n"
}

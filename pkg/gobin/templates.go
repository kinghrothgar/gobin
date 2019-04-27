package gobin

import (
	"bytes"
	"errors"
	htmlTemplate "html/template"
	textTemplate "text/template"

	"github.com/levenlabs/errctx"
)

type Tabs struct {
	Home bool
	Form bool
	Top  bool
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
	Domain  string
	Scheme  string
	Title   string
	ID      string
	Secret string
	Tabs    *Tabs
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

type Templates struct {
	html   *htmlTemplate.Template
	text   *textTemplate.Template
	domain string
	title  string
}

func unescaped(x string) interface{} {
	return htmlTemplate.HTML(x)
}

func NewTemplates(htmlTmplsPath string, textTmplsPath string, domain string) (*Templates, error) {
	htmlTmpls, err := htmlTemplate.ParseFiles(htmlTmplsPath)
	if err != nil {
		return nil, errctx.Mark(err)
	}
	textTmpls, err := textTemplate.ParseFiles(textTmplsPath)
	if err != nil {
		return nil, errctx.Mark(err)
	}
	return &Templates{
		html:   htmlTmpls,
		text:   textTmpls,
		domain: domain,
	}, nil
}

func (t *Templates) GetHomePage(contentType string) ([]byte, error) {
	tabs := &Tabs{Home: true}
	page := &HomePage{Domain: t.domain, Title: t.title, Tabs: tabs}
	return t.execute(contentType, "homePage", page)
}

func (t *Templates) GetMessPage(contentType string, message string) ([]byte, error) {
	tabs := &Tabs{}
	page := &MessPage{Title: t.title, Tabs: tabs, Message: message}
	return t.execute(contentType, "messPage", page)
}

func (t *Templates) GetURLPage(scheme, contentType, id, secret string) ([]byte, error) {
	tabs := &Tabs{Form: true}
	page := &URLPage{
		Domain:  t.domain,
		Scheme:  scheme,
		Title:   t.title,
		ID:      id,
		Secret: secret,
		Tabs:    tabs,
	}
	return t.execute(contentType, "urlPage", page)
}

// BuildURLs builds the urls given the scheme (http/https), id and secret
func (t *Templates) BuildURLs(scheme, id, secret string) string {
	urls := scheme + "://" + t.domain + "/" + id + "\n"
	urls += scheme + "://" + t.domain + "/expire/" + secret + "\n"
	return urls
}

func (t *Templates) execute(contentType string, tmplName string, data interface{}) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}
	switch contentType {
	case "HTML":
		err = t.html.ExecuteTemplate(buf, tmplName, data)
		break
	case "TEXT":
		err = t.text.ExecuteTemplate(buf, tmplName, data)
		break
	default:
		err = errctx.Mark(errors.New("invalid content type"))
	}
	return buf.Bytes(), err
}

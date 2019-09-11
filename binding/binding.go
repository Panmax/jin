package binding

import "net/http"

const (
	MIMEJSON = "application/json"
	MIMEHTML = "text/html"
	MIMEXML  = "application/xml"
	MIMEXML2 = "text/xml"
)

type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

var (
	JSON = jsonBinding{}
	XML  = xmlBinding{}
	Form = formBinding{}
)

func Default(method, contentType string) Binding {
	if method == "GET" {
		return Form
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	}
}

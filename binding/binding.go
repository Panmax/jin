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

type BindingBody interface {
	Binding
	BindBody([]byte, interface{}) error
}

type BindingUri interface {
	Name() string
	BindUri(map[string][]string, interface{}) error
}

type StructValidator interface {
	ValidateStruct(interface{}) error
	Engine() interface{}
}

var Validator StructValidator = &defaultValidator{}

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

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}

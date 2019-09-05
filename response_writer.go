package jin

import "net/http"

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	Status() int

	Size() int

	WriteString(string) (int, error)

	Written() bool

	WriteHeaderNow()

	Pusher() http.Pusher
}

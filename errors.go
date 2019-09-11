package jin

type ErrorType uint64

const (
	ErrorTypeBind    ErrorType = 1 << 63
	ErrorTypeRender  ErrorType = 1 << 62
	ErrorTypePrivate ErrorType = 1 << 0
	ErrorTypePublic  ErrorType = 1 << 1
	ErrorTypeAny     ErrorType = 1<<64 - 1
	ErrorTypeNu      ErrorType = 2
)

type Error struct {
	Err  error
	Type ErrorType
	Meta interface{}
}

type errorMsgs []*Error

var _ error = &Error{}

func (msg *Error) SetType(flags ErrorType) *Error {
	msg.Type = flags
	return msg
}

func (msg *Error) Error() string {
	return msg.Err.Error()
}

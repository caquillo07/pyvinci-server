package server

type PublicError interface {
	PublicError() string
	Code() int
}

type publicError struct {
	msg  string
	code int
}

func newPublicError(msg string, code int) error {
	return publicError{msg: msg, code: code}
}

func (e publicError) Error() string {
	return e.msg
}

func (e publicError) Code() int {
	return e.code
}

func (e publicError) PublicError() string {
	return e.Error()
}

func newValidationError(msg string) error {
	return publicError{msg: msg, code: 400}
}

func newNotFoundError(msg string) error {
	return publicError{msg: msg, code: 404}
}

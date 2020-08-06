package server

type publicError interface {
    PublicError() string
    Code() int
}

type validationError struct {
    msg string
}

func newValidationError(msg string) error {
    return validationError{msg: msg}
}

func (e validationError) Error() string {
    return e.msg
}

func (e validationError) Code() int {
    return 401
}

func (e validationError) PublicError() string {
    return e.Error()
}

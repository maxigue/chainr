package httputil

type ErrorWithStatus interface {
	Error() string
	Status() int
}

type ews struct {
	err    error
	status int
}

func (e *ews) Error() string {
	return e.err.Error()
}

func (e *ews) Status() int {
	return e.status
}

func NewErrorWithStatus(err error, status int) ErrorWithStatus {
	return &ews{err, status}
}

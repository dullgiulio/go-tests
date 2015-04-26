package sima

import "fmt"

type errorCode string

const (
	errorCodeConnFailed errorCode = "err-connection-failed"
)

type codedError struct {
	name errorCode
	err  string
}

func (p *codedError) Error() string {
	return fmt.Sprintf("%s: %s", string(p.name), p.err)
}

type ErrConnectionFailed codedError

func NewErrConnectionFailed(e string) error {
	err := &ErrConnectionFailed{errorCodeConnFailed, e}
	return error(err)
}

func (e *ErrConnectionFailed) Error() string {
	return (*codedError)(e).Error()
}

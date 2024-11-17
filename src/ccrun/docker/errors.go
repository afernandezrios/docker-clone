package docker

import (
	"errors"
	"net/http"
)

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInternalError  = errors.New("internal service error")
	ErrNotImplemented = errors.New("service method not implemented")
)

type UnauthorizedError struct {
	err      error
	authInfo string
}

func NewUnAuthorizedError(response http.Response) error {
	return UnauthorizedError{
		err:      ErrUnauthorized,
		authInfo: response.Header.Get("www-authenticate"),
	}
}

func (e UnauthorizedError) Error() string {
	return e.err.Error()
}

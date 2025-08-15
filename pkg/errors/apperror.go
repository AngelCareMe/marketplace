package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrInternal = errors.New("internal server error")
)

type AppError struct {
	code    string
	message string
	error   error
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		code:    code,
		message: message,
		error:   err,
	}
}

func (a *AppError) Error() string {
	if a == nil {
		return "<nil>"
	}
	if a.error != nil {
		return fmt.Sprintf("[%s] %s: %v", a.code, a.message, a.error)
	}
	return fmt.Sprintf("[%s] %s", a.code, a.message)
}

func (a *AppError) Unwrap() error { return a.error }

func (a *AppError) Code() string { return a.code }

func (a *AppError) Message() string { return a.message }

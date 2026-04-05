package errors

import (
	stderrors "errors"
	"net/http"
	"runtime"
)

type Code string

const (
	CodeInternal         Code = "InternalError"
	CodeResourceNotFound Code = "ResourceNotFound"
	CodeBadRequest       Code = "BadRequest"
	CodeForbidden        Code = "Forbidden"
	CodeUnauthorized     Code = "Unauthorized"
	CodeConflict         Code = "Conflict"
)

type Frame struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

type Error struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`

	Err   error   `json:"-"`
	Stack []Frame `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewTechnical(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if stderrors.As(err, &appErr) {
		if len(appErr.Stack) == 0 {
			appErr.Stack = captureStack(3)
		}
		return appErr
	}

	return &Error{
		Code:       CodeInternal,
		Message:    "internal error",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
		Stack:      captureStack(3),
	}
}

func NewBusiness(code Code, message string, status int, err error) *Error {
	if err == nil {
		return &Error{
			Code:       code,
			Message:    message,
			HTTPStatus: status,
			Stack:      captureStack(3),
		}
	}

	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Err:        err,
		Stack:      captureStack(3),
	}
}

func Internal(err error) *Error {
	return NewBusiness(CodeInternal, "Внутренняя ошибка сервиса", http.StatusInternalServerError, err)
}

func NotFound(message string, err error) *Error {
	return NewBusiness(CodeResourceNotFound, message, http.StatusNotFound, err)
}

func BadRequest(message string, err error) *Error {
	return NewBusiness(CodeBadRequest, message, http.StatusBadRequest, err)
}

func Forbidden(message string, err error) *Error {
	return NewBusiness(CodeForbidden, message, http.StatusForbidden, err)
}

func Unauthorized(message string, err error) *Error {
	return NewBusiness(CodeUnauthorized, message, http.StatusUnauthorized, err)
}

func Conflict(message string, err error) *Error {
	return NewBusiness(CodeConflict, message, http.StatusConflict, err)
}

func captureStack(skip int) []Frame {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)

	stack := make([]Frame, 0, n)
	for {
		frame, more := frames.Next()
		stack = append(stack, Frame{
			Func: frame.Function,
			File: frame.File,
			Line: frame.Line,
		})
		if !more {
			break
		}
	}

	return stack
}
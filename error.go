package gopetri

import (
	"fmt"
)

type errorCode int

const (
	ErrCodePlaceAlreadyRegistered errorCode = iota + 1
	ErrCodeTransitionAlreadyRegistered
	ErrCodePlaceIsNotRegistered
	ErrCodeUnexpectedAvailableTransitionsNumber
	ErrCodeHasNotChipForNewPlace
	ErrCodeNetInErrState
	ErrCodeBeforePlaceReturnedErr
	ErrCodeBeforeTransitReturnedErr
)

type Error struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

func NewError(errorCode errorCode, msg string) *Error {
	return &Error{
		Code:    errorCode,
		Message: msg,
	}
}

func NewErrorf(errorCode errorCode, format string, args ...interface{}) *Error {
	return NewError(errorCode, fmt.Sprintf(format, args...))
}

func (e *Error) Is(errorCode errorCode) bool {
	return e.Code == errorCode
}

func (e *Error) Error() string {
	return e.Message
}

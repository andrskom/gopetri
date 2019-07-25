package gopetri

import (
	"fmt"
)

type ErrCode int

const (
	ErrCodePlaceAlreadyRegistered ErrCode = iota + 1
	ErrCodeTransitionAlreadyRegistered
	ErrCodePlaceIsNotRegistered
	ErrCodeUnexpectedAvailableTransitionsNumber
	ErrCodeHasNotChipForNewPlace
	ErrCodeNetInErrState
	ErrCodeBeforePlaceReturnedErr
	ErrCodeBeforeTransitReturnedErr
	ErrCodeCantSetErrState
	ErrCodeFinished
	ErrCodeFromTransitionAlreadyRegistered
	ErrCodeToTransitionAlreadyRegistered
	ErrCodeConsumerNotSet
	ErrCodePoolAlreadyInit
	ErrCodeWaitingForNetFromPoolTooLong
)

// Error is err model of component.
type Error struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

// NewError init error.
func NewError(errorCode ErrCode, msg string) *Error {
	return &Error{
		Code:    errorCode,
		Message: msg,
	}
}

// NewErrorf init error by format.
func NewErrorf(errorCode ErrCode, format string, args ...interface{}) *Error {
	return NewError(errorCode, fmt.Sprintf(format, args...))
}

// Is check error.
func (e *Error) Is(errorCode ErrCode) bool {
	return e.Code == errorCode
}

func (e *Error) Error() string {
	return e.Message
}

// Is compare err and error code. If err is not nil, is *Error type and have the same code or false.
func Is(err error, code ErrCode) bool {
	if err == nil {
		return false
	}
	error, ok := err.(*Error)
	if !ok {
		return false
	}
	return error.Code == code
}

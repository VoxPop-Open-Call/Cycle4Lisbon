package httputil

import (
	"fmt"
	"net/http"
)

// Error type contains the http status code, the error code and a custom
// message with the details.
//
// Implements the error interface.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code" example:"Record Not Found"`
	Message string `json:"message" example:"user not found"`
}

func (e Error) Error() string {
	return fmt.Sprintf("ApiError{code: %s, message: %s}",
		e.Code,
		e.Message,
	)
}

type ErrorCode struct {
	value  string
	status int
}

var (
	InternalServerError = ErrorCode{
		"Internal Server Error",
		http.StatusInternalServerError,
	}
	MissingAuthToken = ErrorCode{
		"Missing Authorization Token",
		http.StatusUnauthorized,
	}
	InvalidAuthToken = ErrorCode{
		"Invalid Authorization Token",
		http.StatusUnauthorized,
	}
	TokenUserNotFound = ErrorCode{
		"Token User Not Found",
		http.StatusUnauthorized,
	}
	Forbidden = ErrorCode{
		"Forbidden Action",
		http.StatusForbidden,
	}
	AdminAccessRequired = ErrorCode{
		"Admin Access Required",
		http.StatusForbidden,
	}
	BadRequest = ErrorCode{
		"Bad Request",
		http.StatusBadRequest,
	}
	InvalidUUID = ErrorCode{
		"Invalid UUID",
		http.StatusBadRequest,
	}
	RecordNotFound = ErrorCode{
		"Record Not Found",
		http.StatusNotFound,
	}
	EmailAlreadyRegistered = ErrorCode{
		"Email Already Registered",
		http.StatusBadRequest,
	}
	UsernameAlreadyRegistered = ErrorCode{
		"Username Already Registered",
		http.StatusBadRequest,
	}
	PasswordResetCodeExpired = ErrorCode{
		"Password Reset Code Expired",
		http.StatusGone,
	}
	PasswordResetCodeUsed = ErrorCode{
		"Password Reset Code Already Used",
		http.StatusGone,
	}
	IncorrectPassword = ErrorCode{
		"Incorrect Password",
		http.StatusForbidden,
	}
	InvalidFile = ErrorCode{
		"Invalid File",
		http.StatusBadRequest,
	}
	InvalidGPXFile = ErrorCode{
		"Invalid GPX File",
		http.StatusBadRequest,
	}
	DuplicatedGPXFile = ErrorCode{
		"Duplicated GPX File",
		http.StatusBadRequest,
	}
	ImportReadError = ErrorCode{
		"Import Read Error",
		http.StatusBadRequest,
	}
	ImportCSVMissingHeader = ErrorCode{
		"Import CSV Missing Header",
		http.StatusBadRequest,
	}
	ImportMissingColumn = ErrorCode{
		"Import Missing Column",
		http.StatusBadRequest,
	}
	ImportInvalidValue = ErrorCode{
		"Import Invalid Value",
		http.StatusBadRequest,
	}
)

const (
	ForbiddenMessage     = "the user does not have access to this operation"
	AdminRequiredMessage = "the user must be an administrator to perform this action"
)

// NewErrorMsg creates an httputil.Error with given error code and message.
func NewErrorMsg(code ErrorCode, msg string) Error {
	return Error{code.status, code.value, msg}
}

// NewError creates an httputil.Error with given error code and err's message.
func NewError(code ErrorCode, err error) Error {
	return NewErrorMsg(code, err.Error())
}

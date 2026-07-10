// Package apperror 定义跨应用入口共享的稳定错误契约。
package apperror

import (
	"errors"
	"fmt"
)

// Code 是调用方可稳定判断的错误码。
type Code string

const (
	CodeInvalidArgument Code = "invalid_argument"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeUnauthorized    Code = "unauthorized"
	CodeForbidden       Code = "forbidden"
	CodeResourceLimit   Code = "resource_limit"
	CodeUnavailable     Code = "unavailable"
	CodeInternal        Code = "internal"
)

// Error 同时保存稳定错误码、可公开消息和原始错误。
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
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

// New 创建不包含底层错误的应用错误。
func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

// Wrap 为底层错误附加稳定错误码和公开消息。
func Wrap(code Code, message string, err error) error {
	if err == nil {
		return New(code, message)
	}
	return &Error{Code: code, Message: message, Err: err}
}

// CodeOf 读取应用错误码；未知错误统一视为内部错误。
func CodeOf(err error) Code {
	var appErr *Error
	if errors.As(err, &appErr) && appErr.Code != "" {
		return appErr.Code
	}
	return CodeInternal
}

// PublicMessage 返回可以安全暴露给调用方的消息。
func PublicMessage(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) && appErr.Message != "" {
		return appErr.Message
	}
	return "internal server error"
}

// IsCode 判断错误链是否包含指定应用错误码。
func IsCode(err error, code Code) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == code
}

// Errorf 创建带格式化公开消息的应用错误。
func Errorf(code Code, format string, args ...any) error {
	return New(code, fmt.Sprintf(format, args...))
}

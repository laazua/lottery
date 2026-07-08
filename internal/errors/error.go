// Package errors 提供带错误码的业务错误类型和预定义错误常量。
package errors

import "fmt"

// Error 是带错误码的业务错误类型。
// Code 为层级编码（如 "E1NT001"），Message 为人类可读描述。
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is 支持 errors.Is 通过错误码匹配。
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewError 创建带错误码的业务错误。
func NewError(code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

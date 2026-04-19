package errors

import "fmt"

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误码
	ErrCodeSuccess         ErrorCode = 0
	ErrCodeInternal        ErrorCode = 10001
	ErrCodeInvalidParams   ErrorCode = 10002
	ErrCodeUnauthorized    ErrorCode = 10003
	ErrCodeForbidden       ErrorCode = 10004
	ErrCodeNotFound        ErrorCode = 10005
	ErrCodeTooManyRequests ErrorCode = 10006

	// API Key相关错误码
	ErrCodeAPIKeyInvalid  ErrorCode = 20001
	ErrCodeAPIKeyExpired  ErrorCode = 20002
	ErrCodeAPIKeyDisabled ErrorCode = 20003

	// 配额相关错误码
	ErrCodeQuotaExhausted ErrorCode = 30001

	// LLM服务相关错误码
	ErrCodeModelNotSupported ErrorCode = 40001
	ErrCodeLLMServiceError   ErrorCode = 40002
	ErrCodeLLMTimeout        ErrorCode = 40003
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf 创建带格式化的错误
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装已有错误
func Wrap(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 实现errors.Unwrap接口
func (e *Error) Unwrap() error {
	return e.Err
}

// Is 判断错误类型
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

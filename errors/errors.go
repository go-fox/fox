package errors

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	foxStatus "github.com/go-fox/fox/api/gen/go/status"

	"github.com/go-fox/fox/internal/bytesconv"
)

var _ error = (*Error)(nil)

// SupportPackageIsVersion1 generate
const SupportPackageIsVersion1 = true

const (
	UnknownCode   = 500 // UnknownCode 未知错误
	UnknownReason = ""  // UnknownReason 未知错误原因
)

// Error 错误定义
type Error struct {
	foxStatus.Status
	cause error
	stack *stack
}

// Error impl error.Error
func (e *Error) Error() string {
	if e.stack != nil {
		return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v stack= %+v", e.Code, e.Reason, e.Message, e.Metadata, e.cause, e.stack)
	}
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v ", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

// Stack get error stack
func (e *Error) Stack() string {
	return fmt.Sprintf("%+v\n", e.stack)
}

// JSON json data
func (e *Error) JSON() string {
	v, _ := json.Marshal(e)
	return bytesconv.BytesToString(v)
}

// Unwrap get Error.cause
func (e *Error) Unwrap() error { return e.cause }

// Is verifying code and Reason
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

// WithCause setting cause
func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

// WithStack setting error stack
func (e *Error) WithStack() *Error {
	e.stack = callers()
	return e
}

// WithMetadata setting metadata
func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md
	return err
}

// GRPCStatus get grpc status
func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(codes.Code(e.Code), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Reason,
			Metadata: e.Metadata,
		})
	return s
}

// New create an error
func New(code int, reason, message string) *Error {
	err := &Error{
		Status: foxStatus.Status{
			Code:    int32(code),
			Message: message,
			Reason:  reason,
		},
	}
	return err.WithStack()
}

// Errorf is new a error
func Errorf(code int, reason, format string, a ...interface{}) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

// Code get Error.code
func Code(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}
	return int(FromError(err).Code)
}

// Reason 错误原因
func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

// Clone 克隆错误
func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		cause: err.cause,
		Status: foxStatus.Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
	}
}

// FromError 从error构造*Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if !ok {
		return New(UnknownCode, UnknownReason, err.Error())
	}
	ret := New(
		int(gs.Code()),
		UnknownReason,
		gs.Message(),
	)
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Reason = d.Reason
			return ret.WithMetadata(d.Metadata)
		default:
		}
	}
	ret.cause = err
	return ret.WithStack()
}

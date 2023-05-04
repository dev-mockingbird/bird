package bird

import (
	"github.com/dev-mockingbird/errors"
)

const (
	CodeOK               = "ok"
	CodeInvalidArguments = "invalid-arguments"
	CodeUnkownError      = "unknown"
	CodeBadFormat        = "bad-format"
	CodeUnauthorized     = "unauthorized"
)

type ResponseBody struct {
	Code string `json:"code"`
	Data any    `json:"data,omitempty"`
}

type Message struct {
	Msg string `json:"msg,omitempty"`
}

func Msg(msg string) Message {
	return Message{Msg: msg}
}

func OK(data any) ResponseBody {
	return ResponseBody{
		Code: CodeOK,
		Data: data,
	}
}

func InvalidArguments(err error, msg ...string) ResponseBody {
	return ErrorOccurred(err, CodeInvalidArguments, msg...)
}

func UnknownError(err error, msg ...string) ResponseBody {
	return ErrorOccurred(err, CodeUnkownError, msg...)
}

func Unauthorized(err error, msg ...string) ResponseBody {
	return ErrorOccurred(err, CodeUnauthorized, msg...)
}

// parse err tag and msg.
// for untagged err, it can't produce the err detail as message for client
// only the tagged most ancient ancestor error can produce client message and code
func ErrorOccurred(err error, defaultcode string, msgs ...string) ResponseBody {
	err = errors.Ancestor(err)
	codes, msg := errors.Parse(err)
	code := defaultcode
	if len(codes) > 0 {
		code = codes[0]
	}
	if msg == "" && len(msgs) > 0 {
		msg = msgs[0]
	}
	return ResponseBody{
		Code: code,
		Data: Msg(msg),
	}
}

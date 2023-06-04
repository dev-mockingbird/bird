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
	if err = errors.LastTagged(err); err == nil {
		return ResponseBody{Code: defaultcode, Data: Msg(func() string {
			if len(msgs) > 0 {
				return msgs[0]
			}
			return defaultcode
		}())}
	}
	code := defaultcode
	if tags := errors.Tags(err); len(tags) > 0 {
		code = tags[0]
	}
	msg := code
	if e := errors.Unwrap(err); e != nil {
		msg = e.Error()
	}
	return ResponseBody{
		Code: code,
		Data: Msg(msg),
	}
}

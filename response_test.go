package bird

import (
	"testing"

	"github.com/dev-mockingbird/errors"
)

func assertResponseBody(t *testing.T, body ResponseBody, code string, msg string) {
	if body.Code != code {
		t.Fatal("code error")
	}
	m, ok := body.Data.(Message)
	if !ok {
		t.Fatal("data type error")
	}
	if m.Msg != msg {
		t.Fatal("msg error")
	}
}

func TestErrorOccurred(t *testing.T) {
	respBody := ErrorOccurred(errors.New("hello world"), "default-code")
	assertResponseBody(t, respBody, "default-code", "default-code")
	respBody = ErrorOccurred(errors.New("hello world", "has-code"), "default-code")
	assertResponseBody(t, respBody, "has-code", "hello world")
	respBody = ErrorOccurred(errors.New("hello world"), "default-code", "default msg")
	assertResponseBody(t, respBody, "default-code", "default msg")
}

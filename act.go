package bird

import (
	"net/http"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/validate"
)

type HandleFunc func(Actor)

// sample:
//
//	r := GinRouter(engine.RouterGroup, logger.New())
//	r.ON("/hello-world", func(actor Actor) {
//	    actor.Write(http.StatusOK, "hello world! ^_^")
//	}).Prepare()
type Actor interface {
	Bind(data any) error
	RequestId() string
	Set(key string, data any)
	Get(key string) (any, bool)
	Query(key string) string
	QueryArray(key string) []string
	Param(key string) string
	Next()
	GetRequest() *http.Request
	GetResponseWriter() http.ResponseWriter
	Validate(data any, rules ...validate.Rules) error
	Write(statusCode int, data any)
	Logger() logf.Logger
}

type Entry interface {
	Prepare(methods ...string)
}

type Router interface {
	Use(...HandleFunc)
	Group(string) Router
	ON(path string, act ...HandleFunc) Entry
	HttpHandler() http.Handler
}

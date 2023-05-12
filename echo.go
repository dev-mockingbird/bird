package bird

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/validate"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type echoEntry struct {
	g      *echo.Echo
	logger logf.Logger
	path   string
	acts   []HandleFunc
}

func constructEchoActor(ctx echo.Context, logger logf.Logger) Actor {
	reqId := ctx.Request().Header.Get("Request-Id")
	if reqId == "" {
		reqId = uuid.NewString()
		ctx.Request().Header.Add("Request-Id", reqId)
	}
	method := strings.ToUpper(ctx.Request().Method)
	path := ctx.Request().URL.Path
	actor := EchoActor(ctx, logger.Prefix(fmt.Sprintf("%s %s[%s]: ", method, path, reqId)))
	return actor
}

func (entry echoEntry) Prepare(methods ...string) {
	for _, act := range entry.acts {
		h := func(ctx echo.Context) error {
			act(constructEchoActor(ctx, entry.logger))
			return nil
		}
		if len(methods) == 0 {
			entry.g.Any(entry.path, h)
			return
		}
		entry.g.Match(methods, entry.path, h)
	}
}

type echoActor struct {
	echo.Context
	next      echo.HandlerFunc
	logger    logf.Logger
	validator validate.Validator
}

type echoRouter struct {
	logger logf.Logger
	g      *echo.Echo
}

var _ Router = &echoRouter{}

func EchoRouter(r *echo.Echo, logger logf.Logger) Router {
	return &echoRouter{logger: logger, g: r}
}

func (r echoRouter) Use(acts ...HandleFunc) {
	for _, act := range acts {
		r.g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				act(constructEchoActor(ctx, r.logger))
				return nil
			}
		})
	}
}

func (r echoRouter) ON(path string, acts ...HandleFunc) Entry {
	return echoEntry{
		path:   path,
		logger: r.logger,
		g:      r.g,
		acts:   acts,
	}
}

func (g echoRouter) HttpHandler() http.Handler {
	return g.g
}

var _ Actor = &echoActor{}

func EchoActor(ctx echo.Context, logger logf.Logger) *echoActor {
	return &echoActor{
		Context:   ctx,
		logger:    logger,
		validator: validate.GetValidator(validate.Logger(logger)),
	}
}

func (g echoActor) Validate(data any, rules ...validate.Rules) error {
	if err := g.validator.Validate(data, rules...); err != nil {
		g.logger.Logf(logf.Error, "validate: %s", err.Error())
		return err
	}
	return nil
}

func (g echoActor) Query(key string) string {
	return g.Context.QueryParam(key)
}

func (g echoActor) Next() {
	if g.next != nil {
		g.next(g.Context)
	}
}

func (g echoActor) QueryArray(key string) []string {
	return g.Context.QueryParams()[key]
}

func (g echoActor) Logger() logf.Logger {
	return g.logger
}

func (g echoActor) Set(key string, data any) {
	g.Context.Set(key, data)
}

func (g echoActor) Get(key string) (any, bool) {
	ret := g.Context.Get(key)
	if ret == nil {
		return ret, false
	}
	return ret, true
}

func (g echoActor) Write(statusCode int, data any) {
	g.Context.Request().Header.Add("Request-Id", g.RequestId())
	g.Context.JSON(statusCode, data)
}

func (g echoActor) GetRequest() *http.Request {
	return g.Context.Request()
}

func (g echoActor) RequestId() string {
	return g.Context.Request().Header.Get("Request-Id")
}

func (g echoActor) GetResponseWriter() http.ResponseWriter {
	return g.Context.Response().Writer
}

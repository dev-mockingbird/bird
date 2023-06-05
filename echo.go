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
	g      *echo.Group
	logger logf.Logger
	path   string
	acts   []HandleFunc
}

type EchoContextGetter interface {
	GetContext() echo.Context
}

func constructEchoActor(ctx echo.Context, logger logf.Logger, next echo.HandlerFunc) Actor {
	reqId := ctx.Request().Header.Get("Request-Id")
	if reqId == "" {
		reqId = uuid.NewString()
		ctx.Request().Header.Add("Request-Id", reqId)
	}
	method := strings.ToUpper(ctx.Request().Method)
	path := ctx.Request().URL.Path
	actor := EchoActor(ctx, logger.Prefix(fmt.Sprintf("%s %s[%s]: ", method, path, reqId)), next)
	return actor
}

func (entry echoEntry) Prepare(methods ...string) {
	for _, act := range entry.acts {
		h := func(ctx echo.Context) error {
			act(constructEchoActor(ctx, entry.logger, nil))
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
	e      *echo.Echo
	g      *echo.Group
}

var _ Router = &echoRouter{}

func EchoRouter(e *echo.Echo, logger logf.Logger) Router {
	return &echoRouter{logger: logger, e: e, g: e.Group("")}
}

func (r echoRouter) Use(acts ...HandleFunc) {
	for _, act := range acts {
		r.g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				actor := constructEchoActor(ctx, r.logger, next)
				act(actor)
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

func (r echoRouter) Group(base string) Router {
	return echoRouter{
		logger: r.logger.Prefix(base + ": "),
		e:      r.e,
		g:      r.e.Group(base),
	}
}

func (g echoRouter) HttpHandler() http.Handler {
	return g.e
}

var _ Actor = &echoActor{}

func EchoActor(ctx echo.Context, logger logf.Logger, next ...echo.HandlerFunc) *echoActor {
	return &echoActor{
		Context:   ctx,
		logger:    logger,
		validator: validate.GetValidator(validate.Logger(logger)),
		next: func() echo.HandlerFunc {
			if len(next) > 0 {
				return next[0]
			}
			return nil
		}(),
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

func (g echoActor) Write(statusCode int, data any) error {
	g.Context.Request().Header.Add("Request-Id", g.RequestId())
	return g.Context.JSON(statusCode, data)
}

func (g echoActor) GetRequest() *http.Request {
	return g.Context.Request()
}

func (g echoActor) RequestId() string {
	return g.Context.Request().Header.Get("Request-Id")
}

func (g echoActor) GetContext() echo.Context {
	return g.Context
}

func (g echoActor) GetResponseWriter() http.ResponseWriter {
	return g.Context.Response().Writer
}

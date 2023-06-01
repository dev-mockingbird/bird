package bird

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/validate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ginEntry struct {
	g      *gin.RouterGroup
	logger logf.Logger
	path   string
	acts   []HandleFunc
}

type GinContextGetter interface {
	GetContext() *gin.Context
}

func constructGinActor(ctx *gin.Context, logger logf.Logger) Actor {
	reqId := ctx.Request.Header.Get("Request-Id")
	if reqId == "" {
		reqId = uuid.NewString()
		ctx.Request.Header.Add("Request-Id", reqId)
	}
	method := strings.ToUpper(ctx.Request.Method)
	path := ctx.Request.URL.Path
	actor := GinActor(ctx, logger.Prefix(fmt.Sprintf("%s %s [%s]: ", method, path, reqId)))
	return actor
}

func (entry ginEntry) Prepare(methods ...string) {
	ginHandlers := func() []gin.HandlerFunc {
		ret := make([]gin.HandlerFunc, len(entry.acts))
		for i, act := range entry.acts {
			ret[i] = func(ctx *gin.Context) {
				act(constructGinActor(ctx, entry.logger))
			}
		}
		return ret
	}
	if len(methods) == 0 {
		entry.g.Any(entry.path, ginHandlers()...)
		return
	}
	entry.g.Match(methods, entry.path, ginHandlers()...)
}

type ginActor struct {
	*gin.Context
	logger    logf.Logger
	validator validate.Validator
}

type ginRouter struct {
	logger logf.Logger
	r      *gin.RouterGroup
	g      *gin.Engine
}

var _ Router = &ginRouter{}

func GinRouter(g *gin.Engine, logger logf.Logger) Router {
	return &ginRouter{logger: logger, g: g, r: &g.RouterGroup}
}

func (r ginRouter) Use(acts ...HandleFunc) {
	for _, act := range acts {
		r.g.Use(func(ctx *gin.Context) {
			act(constructGinActor(ctx, r.logger))
		})
	}
}

func (r ginRouter) Group(base string) Router {
	return &ginRouter{r: r.g.Group(base), g: r.g, logger: r.logger.Prefix(base + ": ")}
}

func (r ginRouter) ON(path string, acts ...HandleFunc) Entry {
	return ginEntry{
		path:   path,
		logger: r.logger,
		g:      r.r,
		acts:   acts,
	}
}

func (g ginRouter) HttpHandler() http.Handler {
	return g.g
}

var _ Actor = &ginActor{}

func GinActor(ctx *gin.Context, logger logf.Logger) *ginActor {
	return &ginActor{
		Context:   ctx,
		logger:    logger,
		validator: validate.GetValidator(validate.Logger(logger)),
	}
}

func (g ginActor) Validate(data any, rules ...validate.Rules) error {
	if err := g.validator.Validate(data, rules...); err != nil {
		g.logger.Logf(logf.Error, "validate: %s", err.Error())
		return err
	}
	return nil
}

func (g ginActor) Logger() logf.Logger {
	return g.logger
}

func (g ginActor) Set(key string, data any) {
	g.Context.Set(key, data)
}

func (g ginActor) Get(key string) (any, bool) {
	return g.Context.Get(key)
}

func (g ginActor) Bind(obj any) error {
	if err := g.Context.Bind(obj); err != nil {
		g.logger.Logf(logf.Error, "bind object: %s", err.Error())
		return err
	}
	g.logger.Logf(logf.Trace, "get input object: %#v", obj)
	return nil
}

func (g ginActor) Write(statusCode int, data any) {
	g.Header("Request-Id", g.RequestId())
	g.JSON(statusCode, data)
	g.Abort()
}

func (g ginActor) Next() {
	g.Context.Next()
}

func (g ginActor) GetRequest() *http.Request {
	return g.Request
}

func (g ginActor) GetContext() *gin.Context {
	return g.Context
}

func (g ginActor) RequestId() string {
	return g.Request.Header.Get("Request-Id")
}

func (g ginActor) GetResponseWriter() http.ResponseWriter {
	return g.Writer
}

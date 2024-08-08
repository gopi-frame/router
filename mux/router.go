package mux

import (
	pipelinecontract "github.com/gopi-frame/contract/pipeline"
	"github.com/gopi-frame/pipeline"
	"github.com/gopi-frame/response"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	responseconstract "github.com/gopi-frame/contract/response"
	routercontract "github.com/gopi-frame/contract/router"
	"github.com/gopi-frame/exception"
	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router

	middlewares                     []routercontract.Middleware
	middlewareConstructorIndexCache map[reflect.Type]int
	ccType                          reflect.Type
	controllerMethodIndexCache      map[string]int
}

func New() *Router {
	return &Router{
		Router:                          mux.NewRouter(),
		middlewareConstructorIndexCache: make(map[reflect.Type]int),
		controllerMethodIndexCache:      make(map[string]int),
	}
}

func (r *Router) Use(middlewares ...routercontract.Middleware) routercontract.Router {
	if len(middlewares) != 0 {
		r.middlewares = append(r.middlewares, middlewares...)

	}
	return r
}

func (r *Router) Group(group routercontract.RouteGroup, builder func(routercontract.Router)) routercontract.Router {
	g := group.(*RouteGroup)
	g.p = r
	sub := g.Build()
	sub.(*Router).middlewares = append(sub.(*Router).middlewares, g.p.middlewares...)
	builder(sub)
	return sub
}

func (r *Router) Controller(controller routercontract.Controller, builder func(routercontract.Router)) routercontract.Router {
	return r.Group(controller.RouteGroup(), func(r routercontract.Router) {
		if _, ok := controller.(routercontract.ConstructableController); ok {
			r := r.(*Router)
			r.controllerMethodIndexCache = make(map[string]int)
			r.ccType = reflect.TypeOf(controller)
			if r.ccType.Kind() != reflect.Pointer {
				panic(exception.NewArgumentException("controller", controller, "controller should be a pointer"))
			}
			method, _ := r.ccType.MethodByName("Construct")
			r.controllerMethodIndexCache["Construct"] = method.Index
		}
		builder(r)
	})
}

func (r *Router) GET(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodGet}, path, handler)
}

func (r *Router) POST(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodPost}, path, handler)
}

func (r *Router) PUT(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodPut}, path, handler)
}

func (r *Router) PATCH(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodPatch}, path, handler)
}

func (r *Router) DELETE(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodDelete}, path, handler)
}

func (r *Router) OPTIONS(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodOptions}, path, handler)
}

func (r *Router) HEAD(path string, handler routercontract.Handler) routercontract.Route {
	return r.Route([]string{http.MethodHead}, path, handler)
}

func (r *Router) Route(methods []string, path string, handler routercontract.Handler) routercontract.Route {
	if r.ccType != nil {
		fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		ss := strings.Split(fn, ".")
		fn = ss[len(ss)-1]
		isMethod := strings.HasSuffix(fn, "-fm")
		fn = strings.TrimRight(fn, "-fm")
		method, ok := r.ccType.MethodByName(fn)
		if ok && isMethod {
			r.controllerMethodIndexCache[fn] = method.Index
			handler = func(request *http.Request) responseconstract.Responser {
				var cc = reflect.New(r.ccType.Elem())
				cc.Method(r.controllerMethodIndexCache["Construct"]).Call([]reflect.Value{reflect.ValueOf(request)})
				out := cc.Method(r.controllerMethodIndexCache[fn]).Call([]reflect.Value{reflect.ValueOf(request)})
				return out[0].Interface().(responseconstract.Responser)
			}
		}
	}

	route := r.Router.Methods(methods...).Path(path).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp := func(request *http.Request) responseconstract.Responser {
			p := pipeline.New[*http.Request, responseconstract.Responser]().Send(request)
			pipes := make([]pipelinecontract.Pipe[*http.Request, responseconstract.Responser], 0)
			for _, middleware := range r.middlewares {
				if cm, ok := middleware.(routercontract.ConstructableMiddleware); ok {
					cmType := reflect.Indirect(reflect.ValueOf(cm)).Type()
					middleware := reflect.New(cmType)
					var constructor reflect.Value
					if constructorIndex, ok := r.middlewareConstructorIndexCache[cmType]; !ok {
						method, _ := cmType.MethodByName("Construct")
						r.middlewareConstructorIndexCache[cmType] = method.Index
						constructor = middleware.Method(method.Index)
					} else {
						constructor = middleware.Method(constructorIndex)
					}
					constructor.Call([]reflect.Value{
						reflect.ValueOf(request),
					})
					pipes = append(pipes, middleware.Interface().(routercontract.ConstructableMiddleware))
				} else {
					pipes = append(pipes, middleware)
				}
			}
			return p.Through(pipes...).Then(func(request *http.Request) responseconstract.Responser {
				return handler(request)
			})
		}(req)
		if resp == nil {
			resp = response.New(http.StatusOK)
		}
		resp.ServeHTTP(w, req)
	})
	return &Route{
		Route:           route,
		originalHandler: handler,
	}
}

func (r *Router) OnNotFound(handler routercontract.Handler) {
	r.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp := handler(req)
		if resp == nil {
			resp = response.New(http.StatusNotFound)
		}
		resp.ServeHTTP(w, req)
	})
}

func (r *Router) OnMethodNotAllowed(handler routercontract.Handler) {
	r.Router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp := handler(req)
		if resp == nil {
			resp = response.New(http.StatusMethodNotAllowed)
		}
		resp.ServeHTTP(w, req)
	})
}

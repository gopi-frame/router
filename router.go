package router

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	contract "github.com/gopi-frame/contract/pipeline"
	"github.com/gopi-frame/contract/response"
	"github.com/gopi-frame/contract/router"
	"github.com/gopi-frame/exception"
	"github.com/gopi-frame/pipeline"
	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router

	cmcache       map[reflect.Type]int
	ccType        reflect.Type
	ccMethodCache map[string]int
}

func New() *Router {
	return &Router{
		Router:        mux.NewRouter(),
		cmcache:       make(map[reflect.Type]int),
		ccMethodCache: make(map[string]int),
	}
}

func (r *Router) Use(middlewares ...router.Middleware) {
	if len(middlewares) != 0 {
		r.Router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				p := pipeline.New[*http.Request, response.Responser]().Send(req)
				steps := []contract.Pipe[*http.Request, response.Responser]{}
				for _, middleware := range middlewares {
					if cm, ok := middleware.(router.ConstructableMiddleware); ok {
						cmType := reflect.Indirect(reflect.ValueOf(cm)).Type()
						middleware := reflect.New(cmType)
						var constructor reflect.Value
						if constructorIndex, ok := r.cmcache[cmType]; !ok {
							method, _ := cmType.MethodByName("Construct")
							r.cmcache[cmType] = method.Index
							constructor = middleware.Method(method.Index)
						} else {
							constructor = middleware.Method(constructorIndex)
						}
						constructor.Call([]reflect.Value{
							reflect.ValueOf(req),
						})
						steps = append(steps, middleware.Interface().(router.ConstructableMiddleware))
					} else {
						steps = append(steps, middleware)
					}
				}
				resp := p.Through(steps...).Then(func(request *http.Request) response.Responser {
					req = request
					return nil
				})
				if resp != nil {
					next = resp
				}
				next.ServeHTTP(w, req)
			})
		})
	}
}

func (r *Router) Group(group router.RouteGroup, builder func(router.Router)) router.Router {
	g := group.(*RouteGroup)
	g.p = r
	sub := g.Build()
	builder(sub)
	return sub
}

func (r *Router) Controller(controller router.Controller, builder func(router.Router)) router.Router {
	return r.Group(controller.RouterGroup(), func(r router.Router) {
		if _, ok := controller.(router.ConstructableController); ok {
			r := r.(*Router)
			r.ccMethodCache = make(map[string]int)
			r.ccType = reflect.TypeOf(controller)
			if r.ccType.Kind() != reflect.Pointer {
				panic(exception.NewArgumentException("controller", controller, "controller should be a pointer"))
			}
			method, _ := r.ccType.MethodByName("Construct")
			r.ccMethodCache["Construct"] = method.Index
		}
		builder(r)
	})
}

func (r *Router) GET(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodGet}, path, handler)
}

func (r *Router) POST(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodPost}, path, handler)
}

func (r *Router) PUT(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodPut}, path, handler)
}

func (r *Router) PATCH(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodPatch}, path, handler)
}

func (r *Router) DELETE(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodDelete}, path, handler)
}

func (r *Router) OPTIONS(path string, handler func(request *http.Request) response.Responser) router.Route {
	return r.Route([]string{http.MethodOptions}, path, handler)
}

func (r *Router) Route(methods []string, path string, handler func(request *http.Request) response.Responser) router.Route {
	if r.ccType != nil {
		funcname := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		ss := strings.Split(funcname, ".")
		funcname = ss[len(ss)-1]
		isMethod := strings.HasSuffix(funcname, "-fm")
		funcname = strings.TrimRight(funcname, "-fm")
		method, ok := r.ccType.MethodByName(funcname)
		if ok && isMethod {
			r.ccMethodCache[funcname] = method.Index
			handler = func(request *http.Request) response.Responser {
				var cc = reflect.New(r.ccType.Elem())
				cc.Method(r.ccMethodCache["Construct"]).Call([]reflect.Value{reflect.ValueOf(request)})
				out := cc.Method(r.ccMethodCache[funcname]).Call([]reflect.Value{reflect.ValueOf(request)})
				return out[0].Interface().(response.Responser)
			}
		}
	}
	return r.Router.Methods(methods...).Path(path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := handler(r)
		response.ServeHTTP(w, r)
	})
}

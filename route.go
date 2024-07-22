package router

import (
	pipelinecontract "github.com/gopi-frame/contract/pipeline"
	"github.com/gopi-frame/contract/response"
	"github.com/gopi-frame/contract/router"
	"github.com/gopi-frame/pipeline"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
)

type Route struct {
	*mux.Route

	originalHandler                 router.Handler
	middlewareConstructorIndexCache map[reflect.Type]int
	middlewares                     []router.Middleware
}

func (r *Route) Name(name string) router.Route {
	r.Route.Name(name)
	return r
}

func (r *Route) Use(middlewares ...router.Middleware) router.Route {
	if len(middlewares) == 0 {
		return r
	}
	r.middlewares = append(r.middlewares, middlewares...)
	r.Route.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var p = pipeline.New[*http.Request, response.Responser]().Send(req)
		var pipes []pipelinecontract.Pipe[*http.Request, response.Responser]
		for _, middleware := range r.middlewares {
			if cm, ok := middleware.(router.ConstructableMiddleware); ok {
				cmType := reflect.Indirect(reflect.ValueOf(cm)).Type()
				middleware := reflect.New(cmType)
				var constructor reflect.Value
				if constructorIndex, ok := r.middlewareConstructorIndexCache[cmType]; !ok {
					method, _ := cmType.MethodByName("Construct")
					if r.middlewareConstructorIndexCache == nil {
						r.middlewareConstructorIndexCache = make(map[reflect.Type]int)
					}
					r.middlewareConstructorIndexCache[cmType] = method.Index
					constructor = middleware.Method(method.Index)
				} else {
					constructor = middleware.Method(constructorIndex)
				}
				constructor.Call([]reflect.Value{reflect.ValueOf(req)})
				pipes = append(pipes, middleware.Interface().(pipelinecontract.Pipe[*http.Request, response.Responser]))
			} else {
				pipes = append(pipes, middleware)
			}
		}
		resp := p.Through(pipes...).Then(func(request *http.Request) response.Responser {
			req = request
			return r.originalHandler(request)
		})
		resp.ServeHTTP(w, req)
	})
	return r
}

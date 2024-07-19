package router

import (
	"strings"

	"github.com/gopi-frame/contract/router"
	"github.com/gorilla/mux"
)

type RouteGroup struct {
	Prefix string
	Host   string

	p *Router
}

func (r *RouteGroup) Build() router.Router {
	var route *mux.Route
	r.Prefix = strings.TrimSpace(r.Prefix)
	if r.Prefix != "" {
		route = r.p.PathPrefix(r.Prefix)
	}
	r.Host = strings.TrimSpace(r.Host)
	if r.Host != "" {
		if route != nil {
			route = route.Host(r.Host)
		} else {
			route = r.p.Host(r.Host)
		}
	}
	if route != nil {
		return &Router{
			Router: route.Subrouter(),

			cmcache: r.p.cmcache,
		}
	}
	return r.p
}

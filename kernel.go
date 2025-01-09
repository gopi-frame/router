package router

import (
	"net/http"

	"github.com/gopi-frame/contract"

	"github.com/gopi-frame/contract/router"
)

type Kernel struct {
	*http.Server
}

type Option = contract.Option[*http.Server]

type OptionFunc func(srv *http.Server) error

func (f OptionFunc) Apply(srv *http.Server) error {
	return f(srv)
}

func WithAddr(addr string) Option {
	return OptionFunc(func(srv *http.Server) error {
		srv.Addr = addr
		return nil
	})
}

func NewKernel(r router.Router, opts ...Option) (*Kernel, error) {
	srv := &http.Server{
		Handler: r,
	}
	for _, opt := range opts {
		if err := opt.Apply(srv); err != nil {
			return nil, err
		}
	}
	return &Kernel{
		Server: srv,
	}, nil
}

func (k *Kernel) Run() error {
	return k.Server.ListenAndServe()
}

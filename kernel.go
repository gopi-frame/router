package router

import (
	"net/http"

	"github.com/gopi-frame/contract/router"
)

type Kernel struct {
	http.Server
}

func NewKernel(r router.Router) *Kernel {
	return &Kernel{
		http.Server{
			Handler: r,
		},
	}
}

func (k *Kernel) Run() error {
	return k.Server.ListenAndServe()
}

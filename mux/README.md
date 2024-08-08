# Overview
[![Go Reference](https://pkg.go.dev/badge/github.com/gopi-frame/router/mux.svg)](https://pkg.go.dev/github.com/gopi-frame/router/mux)
[![Test mux](https://github.com/gopi-frame/router/actions/workflows/mux.yml/badge.svg)](https://github.com/gopi-frame/router/actions/workflows/mux.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopi-frame/router/mux)](https://goreportcard.com/report/github.com/gopi-frame/router/mux)
[![codecov](https://codecov.io/gh/gopi-frame/router/graph/badge.svg?token=7TV2KL6P6G&flag=mux)](https://codecov.io/gh/gopi-frame/router?flags[0]=mux)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

Package mux provides route handling for HTTP Request base on [gorilla/mux](https://github.com/gorilla/mux)

## Installation
```shell
go get -u -v github.com/gopi-frame/router/mux
```

## Import
```go
import "github.com/gopi-frame/router/mux"
```

## Usage

### Single Route

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
)

var handler = func(r *http.Request) responsecontract.Responser {
    return response.New(http.StatusOK, "Hello World")
}

func main() {
    r := mux.New()
    r.GET("/get", handler)
    r.POST("/post", handler)
    r.PUT("/put", handler)
    r.PATCH("/patch", handler)
    r.DELETE("/delete", handler)
    r.OPTIONS("/options", handler)
    r.Route([]string{http.MethodGet, http.MethodPost}, "/all", handler)
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

### Group Routes

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
)

var handler = func(r *http.Request) responsecontract.Responser {
    return response.New(http.StatusOK, "Hello World")
}

func main() {
    r := mux.New()
    r.Group(&router.RouteGroup{
        Prefix: "/group", // prefix all routes in group
    }, func(r routercontract.Router) {
        r.GET("/get", handler)
        r.POST("/post", handler)
        r.PUT("/put", handler)
        r.PATCH("/patch", handler)
        r.DELETE("/delete", handler)
        r.OPTIONS("/options", handler)
        r.Route([]string{http.MethodGet, http.MethodPost}, "/all", handler)
        // nested group
        r.Group(&router.RouteGroup{
            Prefix: "/nested",
        }, func(r routercontract.Router) {
            // more routes
        })
    })
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

### Subdomain Routes

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
)

var handler = func(r *http.Request) responsecontract.Responser {
    return response.New(http.StatusOK, "Hello World")
}

func main() {
    r := mux.New()
    r.Group(&router.Route    oup{
        Host: "example.com",
    }, func(r routercontract.Router) {
        r.GET("/get", handler)
        r.POST("/post", handler)
        r.PUT("/put", handler)
        r.PATCH("/patch", handler)
        r.DELETE("/delete", handler)
        r.OPTIONS("/options", handler)
        r.Route([]string{http.MethodGet, http.MethodPost}, "/all", handler)
    })
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

### Controller

#### Static controller

Static controller is a set of routes.
It implements `github.com/gopi-frame/contract/router.Controller` interface.
It is used to group routes together.
It is static, so you shall **NEVER** set any active data in the controller's properties.

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
    "net/http"
)

type StaticController struct{}

func (s *StaticController) RouteGroup() routercontract.RouteGroup {
    return &router.RouteGroup{Prefix: "/static"}
}

func (s *StaticController) Get(r *http.Request) responsecontract.Responser {
    return response.New(http.StatusOK, "Hello World")
}

func main() {
    r := mux.New()
    staticController := &StaticController{}
    r.Controller(, func(r routercontract.Router) {
        r.GET("/get", staticController.Get)
    })
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

#### Constructable controller

Constructable controller is a set of routes and also has a `Construct` method.
It implements `github.com/gopi-frame/contract/router.ConstructableController` interface.
Every time a request is made to the route, a new instance of the controller is created and the `Construct` method is
called.
It is useful when you want to initialize a controller with some data.

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
    "net/http"
)

type ConstructableController struct {
    data string
}

func (c *ConstructableController) Construct(r *http.Request) {
    c.data = r.URL.RawQuery
}

func (c *ConstructableController) RouteGroup() routercontract.RouteGroup {
    return &router.RouteGroup{Prefix: "/constructable"}
}

func (c *ConstructableController) Get(r *http.Request) responsecontract.Responser {
    return response.New(http.StatusOK, c.data)
}

func main() {
    r := router.New()
    constructableController := &ConstructableController{}
    r.Controller(constructableController, func(r routercontract.Router) {
        r.GET("/get", constructableController.Get)    
    })
}
```

### Middleware

#### Static middleware

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/router/mux"
    "github.com/gopi-frame/response"
    "net/http"
)

type StaticMiddleware struct{}

func (s *StaticMiddleware) Handle(r *http.Request, next routerconstract.Handler) responsecontract.Responser {
    if r.URL.Path != "/middleware" {
        return response.New(403, "Forbidden")
    }
    return next(r)
}

func main() {
    r := mux.New()
    r.Use(&StaticMiddleware{}) // add global middleware
    r.GET("/middleware", func(r *http.Request) responsecontract.Responser {
        return response.New(http.StatusOK, "Hello World")
    }).Use(&StaticMiddleware{}) // add middleware to specific route
    r.Group(&router.RouteGroup{Prefix: "/group"}, func(r routercontract.Router) {
        r.GET("/get", func(r *http.Request) responsecontract.Responser {
            return response.New(http.StatusOK, "Hello World")
        })
    }).Use(&StaticMiddleware{}) // add middleware to specific group
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

#### Constructable middleware

Constructable middleware is a middleware that implements `github.com/gopi-frame/contract/router.ConstructableMiddleware`
interface.
Every time a request is made to the route, a new instance of the middleware is created and the `Construct` method is Called.
It is useful when you want to initialize a middleware with some data.

```go
package main

import (
    responsecontract "github.com/gopi-frame/contract/response"
    routercontract "github.com/gopi-frame/contract/router"
    "github.com/gopi-frame/router/mux"
    "github.com/gopi-frame/response"
    "net/http"
)

type NonStaticMiddleware struct{
    data string
}

func (s *NonStaticMiddleware) Construct(r *http.Request) {
    s.data = r.URL.RawQuery
}

func (s *NonStaticMiddleware) Handle(r *http.Request, next routerconstract.Handler) responsecontract.Responser {
    if r.URL.Path != "/middleware" {
        return response.New(403, "Forbidden")
    }
    return next(r)
}

func main() {
    r := mux.New()
    r.Use(&NonStaticMiddleware{}) // add global middleware
    r.GET("/middleware", func(r *http.Request) responsecontract.Responser {
        return response.New(http.StatusOK, "Hello World")
    }).Use(&NonStaticMiddleware{}) // add middleware to specific route
    r.Group(&router.RouteGroup{Prefix: "/group"}, func(r routercontract.Router) {
        r.GET("/get", func(r *http.Request) responsecontract.Responser {
            return response.New(http.StatusOK, "Hello World")
        })
    }).Use(&NonStaticMiddleware{}) // add middleware to specific group
    if err := http.ListenAndServe(":8080", r); err != nil {
        panic(err)
    }
}
```

## Custom error handler

### Not Found

```go
package main

import (
    "net/http"
    responsecontract "github.com/gopi-frame/contract/response"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
)

func main() {
    r := mux.New()
    r.NotFound(func(req *http.Request) responsecontract.Responser {
        return response.New(http.StatusNotFound, "Not Found")
    })
}
```

### Method Not Allowed

```go
package main

import (
    "net/http"
    responsecontract "github.com/gopi-frame/contract/response"
    "github.com/gopi-frame/response"
    "github.com/gopi-frame/router/mux"
)

func main() {
    r := mux.New()
    r.MethodNotAllowed(func(req *http.Request) responsecontract.Responser {
        return response.New(http.StatusMethodNotAllowed, "Method Not Allowed")
    })
}
```
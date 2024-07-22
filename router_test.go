package router

import (
	"context"
	responseconstract "github.com/gopi-frame/contract/response"
	routercontract "github.com/gopi-frame/contract/router"
	"github.com/gopi-frame/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type staticController struct {
	Prefix string
	Host   string
}

func (s *staticController) RouteGroup() routercontract.RouteGroup {
	return &RouteGroup{Prefix: s.Prefix, Host: s.Host}
}

func (s *staticController) Get(r *http.Request) responseconstract.Responser {
	return response.New(http.StatusOK).JSON(map[string]any{"prefix": s.Prefix, "host": s.Host})
}

type nonStaticController struct {
	queries url.Values
	Prefix  string
	Host    string
}

func (n *nonStaticController) Construct(r *http.Request) {
	n.queries = r.URL.Query()
}

func (n *nonStaticController) RouteGroup() routercontract.RouteGroup {
	return &RouteGroup{Prefix: n.Prefix, Host: n.Host}
}

func (n *nonStaticController) Get(r *http.Request) responseconstract.Responser {
	return response.New(http.StatusOK).JSON(n.queries)
}

var idKey = struct{}{}

type staticMiddleware struct{}

func (s *staticMiddleware) Handle(r *http.Request, next routercontract.Handler) responseconstract.Responser {
	var resp responseconstract.Responser
	if r.URL.Query().Has("id") {
		ctx := context.WithValue(r.Context(), idKey, r.URL.Query().Get("id"))
		r = r.WithContext(ctx)
		resp = next(r)
	} else {
		resp = response.New(http.StatusForbidden)
	}
	resp.SetHeader("Custom-Header", "After request")
	return resp
}

type nonStaticMiddleware struct {
	method string
}

func (n *nonStaticMiddleware) Construct(req *http.Request) {
	n.method = req.Method
}

func (n *nonStaticMiddleware) Handle(r *http.Request, next routercontract.Handler) responseconstract.Responser {
	var resp responseconstract.Responser
	if r.URL.Query().Has("id") {
		ctx := context.WithValue(r.Context(), idKey, r.URL.Query().Get("id"))
		r = r.WithContext(ctx)
		resp = next(r)
	} else {
		resp = response.New(http.StatusForbidden)
	}
	resp.SetHeader("Request-Method", n.method)
	resp.SetHeader("Custom-Header", "After request")
	return resp
}

func TestRouter_GET(t *testing.T) {
	router := New()
	router.GET("/get", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/get", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/get", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_POST(t *testing.T) {
	router := New()
	router.POST("/post", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/post", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/post", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_PUT(t *testing.T) {
	router := New()
	router.PUT("/put", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/put", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/put", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_PATCH(t *testing.T) {
	router := New()
	router.PATCH("/patch", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/patch", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/patch", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_DELETE(t *testing.T) {
	router := New()
	router.DELETE("/delete", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/delete", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/delete", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_HEAD(t *testing.T) {
	router := New()
	router.HEAD("/head", func(request *http.Request) responseconstract.Responser {
		return nil
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/head", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/head", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_OPTIONS(t *testing.T) {
	router := New()
	router.OPTIONS("/options", func(request *http.Request) responseconstract.Responser {
		return nil
	})
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/options", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/options", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_Route(t *testing.T) {
	router := New()
	router.Route([]string{http.MethodGet, http.MethodPost}, "/multi", func(request *http.Request) responseconstract.Responser {
		return response.New(200, "Hello World!")
	})
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/multi", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "Hello World!", w.Body.String())

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/multi", nil)
		router.ServeHTTP(w2, req2)
		assert.Equal(t, 200, w2.Code)
		assert.Equal(t, "Hello World!", w2.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/multi", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 405, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notfound", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code)
	})
}

func TestRouter_Group(t *testing.T) {
	t.Run("prefix", func(t *testing.T) {
		router := New()
		router.Group(&RouteGroup{Prefix: "/prefix"}, func(router routercontract.Router) {
			router.GET("/get", func(request *http.Request) responseconstract.Responser {
				return response.New(200, "Hello World!")
			})
		})

		t.Run("success", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/prefix/get", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "Hello World!", w.Body.String())
		})

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/get", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 404, w.Code)
		})
	})

	t.Run("subdomain", func(t *testing.T) {
		router := New()
		router.Group(&RouteGroup{Host: "www.example.com"}, func(router routercontract.Router) {
			router.GET("/get", func(request *http.Request) responseconstract.Responser {
				return response.New(200, "Hello World!")
			})
		})

		t.Run("success", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "https://www.example.com/get", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "Hello World!", w.Body.String())
		})

		t.Run("not found", func(t *testing.T) {
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", "https://www.example2.com/get", nil)
			router.ServeHTTP(w2, req2)
			assert.Equal(t, 404, w2.Code)
		})

	})

	t.Run("mix prefix and subdomain", func(t *testing.T) {
		router := New()
		router.Group(&RouteGroup{Prefix: "/prefix", Host: "www.example.com"}, func(router routercontract.Router) {
			router.GET("/get", func(request *http.Request) responseconstract.Responser {
				return response.New(200, "Hello World!")
			})
		})

		t.Run("success", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "https://www.example.com/prefix/get", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "Hello World!", w.Body.String())
		})

		t.Run("method not allowed", func(t *testing.T) {
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("POST", "https://www.example.com/prefix/get", nil)
			router.ServeHTTP(w2, req2)
			assert.Equal(t, 405, w2.Code)
		})

		t.Run("host not matched", func(t *testing.T) {
			w3 := httptest.NewRecorder()
			req3, _ := http.NewRequest("GET", "https://www.example2.com/prefix/get", nil)
			router.ServeHTTP(w3, req3)
			assert.Equal(t, 404, w3.Code)
		})

		t.Run("not found", func(t *testing.T) {
			w4 := httptest.NewRecorder()
			req4, _ := http.NewRequest("GET", "https://www.example.com/get", nil)
			router.ServeHTTP(w4, req4)
			assert.Equal(t, 404, w4.Code)
		})
	})
}

func TestRouter_Controller(t *testing.T) {
	t.Run("static controller", func(t *testing.T) {
		controller := new(staticController)
		controller.Prefix = "/prefix"
		r := New()
		r.Controller(controller, func(router routercontract.Router) {
			router.GET("/get", controller.Get)
		})

		t.Run("success", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/prefix/get", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.JSONEq(t, `{"prefix": "/prefix", "host": ""}`, w.Body.String())
		})

		t.Run("method not allowed", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/prefix/get", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 405, w.Code)
		})

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/notfound", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 404, w.Code)
		})
	})

	t.Run("non-static controller", func(t *testing.T) {
		controller := new(nonStaticController)
		controller.Prefix = "/prefix"
		r := New()
		r.Controller(controller, func(router routercontract.Router) {
			router.GET("/get", controller.Get)
		})
		t.Run("success", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/prefix/get?id=1", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.JSONEq(t, `{"id": ["1"]}`, w.Body.String())
		})

		t.Run("method not allowed", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/prefix/get?id=1", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 405, w.Code)
		})

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/notfound", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 404, w.Code)
		})
	})
}

func TestRouter_Use(t *testing.T) {
	t.Run("static middleware", func(t *testing.T) {
		t.Run("global", func(t *testing.T) {
			r := New()
			r.Use(new(staticMiddleware))
			r.GET("/get", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			r.POST("/post", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			t.Run("pass", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/get?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 200, w.Code)
				assert.JSONEq(t, `{"id": "1"}`, w.Body.String())
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 200, w2.Code)
				assert.JSONEq(t, `{"id": "1"}`, w2.Body.String())
				assert.Equal(t, "After request", w2.Header().Get("Custom-Header"))
			})
			t.Run("block", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/get", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 403, w.Code)
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 403, w2.Code)
				assert.Equal(t, "After request", w2.Header().Get("Custom-Header"))
			})
		})

		t.Run("group", func(t *testing.T) {
			r := New()
			r.Group(&RouteGroup{Prefix: "/group"}, func(router routercontract.Router) {
				router.GET("/get", func(request *http.Request) responseconstract.Responser {
					return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
				})
			}).Use(new(staticMiddleware))
			r.POST("/post", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			t.Run("pass", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/group/get?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 200, w.Code)
				assert.JSONEq(t, `{"id": "1"}`, w.Body.String())
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 200, w2.Code)
				assert.JSONEq(t, `{"id": null}`, w2.Body.String())
				assert.Equal(t, "", w2.Header().Get("Custom-Header"))
			})

			t.Run("block", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/group/get", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 403, w.Code)
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
			})
		})
	})

	t.Run("non-static middleware", func(t *testing.T) {
		t.Run("global", func(t *testing.T) {
			r := New()
			r.Use(new(nonStaticMiddleware))
			r.GET("/get", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			r.POST("/post", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			t.Run("pass", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/get?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 200, w.Code)
				assert.JSONEq(t, `{"id": "1"}`, w.Body.String())
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
				assert.Equal(t, "GET", w.Header().Get("Request-Method"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 200, w2.Code)
				assert.JSONEq(t, `{"id": "1"}`, w2.Body.String())
				assert.Equal(t, "After request", w2.Header().Get("Custom-Header"))
				assert.Equal(t, "POST", w2.Header().Get("Request-Method"))
			})

			t.Run("block", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/get", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 403, w.Code)
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
				assert.Equal(t, "GET", w.Header().Get("Request-Method"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 403, w2.Code)
				assert.Equal(t, "After request", w2.Header().Get("Custom-Header"))
				assert.Equal(t, "POST", w2.Header().Get("Request-Method"))
			})
		})

		t.Run("group", func(t *testing.T) {
			r := New()
			r.Group(&RouteGroup{Prefix: "/group"}, func(router routercontract.Router) {
				router.Use(new(nonStaticMiddleware))
				router.GET("/get", func(request *http.Request) responseconstract.Responser {
					return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
				})
			})
			r.POST("/post", func(request *http.Request) responseconstract.Responser {
				return response.New(200).JSON(map[string]any{"id": request.Context().Value(idKey)})
			})
			t.Run("pass", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/group/get?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 200, w.Code)
				assert.JSONEq(t, `{"id": "1"}`, w.Body.String())
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
				assert.Equal(t, "GET", w.Header().Get("Request-Method"))

				w2 := httptest.NewRecorder()
				req2, err := http.NewRequest("POST", "/post?id=1", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w2, req2)
				assert.Equal(t, 200, w2.Code)
				assert.JSONEq(t, `{"id": null}`, w2.Body.String())
				assert.Equal(t, "", w2.Header().Get("Custom-Header"))
				assert.Equal(t, "", w2.Header().Get("Request-Method"))
			})

			t.Run("block", func(t *testing.T) {
				w := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/group/get", nil)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, 403, w.Code)
				assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
				assert.Equal(t, "GET", w.Header().Get("Request-Method"))
			})
		})
	})
}

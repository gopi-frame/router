package router

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	rc "github.com/gopi-frame/contract/response"
	"github.com/gopi-frame/contract/router"
	"github.com/gopi-frame/response"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testmkey = struct{}{}

type testpassm struct{}

func (m *testpassm) Handle(request *http.Request, next func(*http.Request) rc.Responser) rc.Responser {
	request = request.WithContext(context.WithValue(request.Context(), testmkey, "middleware"))
	return next(request)
}

type testblockm struct{}

func (m *testblockm) Handle(request *http.Request, next func(*http.Request) rc.Responser) rc.Responser {
	return response.New(401).JSON(map[string]any{"msg": "unauthorized"})
}

type testconstructm struct {
	startedAt time.Time
}

func (cm *testconstructm) Construct(request *http.Request) {
	cm.startedAt = time.Now()
}

func (cm *testconstructm) Handle(request *http.Request, next func(*http.Request) rc.Responser) rc.Responser {
	return response.New(401).JSON(map[string]any{"msg": "unauthorized", "startedAt": cm.startedAt.Format("2006-01-02 15:04:05")})
}

type teststatiscc struct {
	page     int
	pageSize int
}

func (c *teststatiscc) RouterGroup() router.RouteGroup {
	return &RouteGroup{
		Prefix: "/static",
	}
}

func (c *teststatiscc) List(request *http.Request) rc.Responser {
	return response.New(200).JSON(map[string]int{
		"page":      c.page,
		"page_size": c.pageSize,
	})
}

type testconstructc struct {
	page     int
	pageSize int
}

func (c *testconstructc) RouterGroup() router.RouteGroup {
	return &RouteGroup{
		Prefix: "/construct",
	}
}

func (c *testconstructc) Construct(request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(request.URL.Query().Get("page_size"))
	c.page = page
	c.pageSize = pageSize
}

func (c *testconstructc) List(request *http.Request) rc.Responser {
	return response.New(200).JSON(map[string]int{
		"page":      c.page,
		"page_size": c.pageSize,
	})
}

func TestRouter(t *testing.T) {
	t.Run("without middleware", func(t *testing.T) {
		r := New()
		r.Group(&RouteGroup{
			Prefix: "/api",
			Host:   "example.com",
		}, func(r router.Router) {
			r.GET("/users", func(request *http.Request) rc.Responser {
				currentRoute := mux.CurrentRoute(request)
				hostTpl, err := currentRoute.GetHostTemplate()
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				return response.New(200).JSON(map[string]any{
					"host": hostTpl,
				})
			})
		})
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		assert.Equal(t, 200, recorder.Result().StatusCode)
		content, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"host": "example.com"}`, string(content))
	})

	t.Run("with middleware and passed", func(t *testing.T) {
		r := New()
		r.Group(&RouteGroup{
			Prefix: "/api",
			Host:   "example.com",
		}, func(r router.Router) {
			r.GET("/users", func(request *http.Request) rc.Responser {
				currentRoute := mux.CurrentRoute(request)
				hostTpl, err := currentRoute.GetHostTemplate()
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				return response.New(200).JSON(map[string]any{
					"host":      hostTpl,
					"ctx_value": request.Context().Value(testmkey),
				})
			})
		}).Use(new(testpassm))
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		assert.Equal(t, 200, recorder.Result().StatusCode)
		content, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"host": "example.com", "ctx_value": "middleware"}`, string(content))
	})

	t.Run("with middleware and blocked", func(t *testing.T) {
		r := New()
		r.Group(&RouteGroup{
			Prefix: "/api",
			Host:   "example.com",
		}, func(r router.Router) {
			r.GET("/users", func(request *http.Request) rc.Responser {
				currentRoute := mux.CurrentRoute(request)
				hostTpl, err := currentRoute.GetHostTemplate()
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				return response.New(200).JSON(map[string]any{
					"host": hostTpl,
				})
			})
		}).Use(new(testblockm))
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		assert.Equal(t, 401, recorder.Result().StatusCode)
		content, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"msg": "unauthorized"}`, string(content))
	})

	t.Run("with constructor middleware", func(t *testing.T) {
		r := New()
		r.Group(&RouteGroup{
			Prefix: "/api",
			Host:   "example.com",
		}, func(r router.Router) {
			r.GET("/users", func(request *http.Request) rc.Responser {
				currentRoute := mux.CurrentRoute(request)
				hostTpl, err := currentRoute.GetHostTemplate()
				if err != nil {
					assert.FailNow(t, err.Error())
				}
				return response.New(200).JSON(map[string]any{
					"host": hostTpl,
				})
			})
		}).Use(new(testconstructm))
		req1 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		recorder1 := httptest.NewRecorder()
		r.ServeHTTP(recorder1, req1)
		assert.Equal(t, 401, recorder1.Result().StatusCode)
		content1, err := io.ReadAll(recorder1.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		req2 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		recorder2 := httptest.NewRecorder()
		r.ServeHTTP(recorder2, req2)
		assert.Equal(t, 401, recorder1.Result().StatusCode)
		content2, err := io.ReadAll(recorder1.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.NotEqual(t, content1, content2)
	})

	t.Run("static controller", func(t *testing.T) {
		sc := new(teststatiscc)
		sc.page = 1
		sc.pageSize = 10

		r := New()
		r.Controller(sc, func(r router.Router) {
			r.GET("/users", sc.List)
		})

		req := httptest.NewRequest(http.MethodGet, "/static/users", nil)
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		assert.Equal(t, 200, recorder.Result().StatusCode)
		content, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"page": 1, "page_size": 10}`, string(content))
	})

	t.Run("constructable controller", func(t *testing.T) {
		cc := new(testconstructc)
		r := New()
		r.Controller(cc, func(r router.Router) {
			r.GET("/users", cc.List)
			r.GET("/users2", func(request *http.Request) rc.Responser {
				return response.New(200).JSON(map[string]any{
					"msg": "ok",
				})
			})
		})

		req1 := httptest.NewRequest(http.MethodGet, "/construct/users?page=2&page_size=20", nil)
		recorder1 := httptest.NewRecorder()
		r.ServeHTTP(recorder1, req1)
		assert.Equal(t, 200, recorder1.Result().StatusCode)
		content, err := io.ReadAll(recorder1.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"page": 2, "page_size": 20}`, string(content))

		req2 := httptest.NewRequest(http.MethodGet, "/construct/users2?page=2&page_size=20", nil)
		recorder2 := httptest.NewRecorder()
		r.ServeHTTP(recorder2, req2)
		assert.Equal(t, 200, recorder2.Result().StatusCode)
		content, err = io.ReadAll(recorder2.Result().Body)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.JSONEq(t, `{"msg": "ok"}`, string(content))
	})
}

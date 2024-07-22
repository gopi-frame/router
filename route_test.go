package router

import (
	responseconstract "github.com/gopi-frame/contract/response"
	"github.com/gopi-frame/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoute_Name(t *testing.T) {
	r := New()
	route := r.GET("/get", func(request *http.Request) responseconstract.Responser {
		return response.New(200).JSON(map[string]any{
			"message": "Hello World",
		})
	})
	route.Name("greeting")
	assert.Equal(t, route.(*Route).Route, r.Router.Get("greeting"))
}

func TestRoute_Use(t *testing.T) {
	t.Run("static middleware", func(t *testing.T) {
		r := New()
		r.GET("/get", func(request *http.Request) responseconstract.Responser {
			return response.New(200).JSON(map[string]any{
				"id": request.Context().Value(idKey),
			})
		}).Use(new(staticMiddleware))

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
		})
	})

	t.Run("non-static middleware", func(t *testing.T) {
		r := New()
		r.GET("/get", func(request *http.Request) responseconstract.Responser {
			return response.New(200).JSON(map[string]any{
				"id": request.Context().Value(idKey),
			})
		}).Use(new(nonStaticMiddleware))

		t.Run("pass", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/get?id=1", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.JSONEq(t, `{"id": "1"}`, w.Body.String())
			assert.Equal(t, "GET", w.Header().Get("Request-Method"))
			assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
		})

		t.Run("block", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/get", nil)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 403, w.Code)
			assert.Equal(t, "GET", w.Header().Get("Request-Method"))
			assert.Equal(t, "After request", w.Header().Get("Custom-Header"))
		})
	})
}

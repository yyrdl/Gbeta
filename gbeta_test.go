//by yyrdl ,MIT License
package gbeta

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func param(ctx *Context, key string) (bool, string) {
	v := ctx.Get(key)
	if v != nil {
		if vv, ok := v.(string); ok {
			return true, vv
		}
	}
	return false, ""
}

func TestRouteMatching(t *testing.T) {
	var app = App()

	app.Get("/v1/path/aticle/123", func(ctx *Context, res Res, req Req) {
		res.WriteHeader(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/v1/path/aticle/123", nil)

	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Response should be OK for %s", "/v1")
	}
}

// middle_ware example
type Exmple_middleware1 int

func (m *Exmple_middleware1) Do(ctx *Context, res Res, req Req, next Next) {
	ctx.Set("name", "jason")
	next(true)
}

type Exmple_middleware2 int

func (m *Exmple_middleware2) Do(ctx *Context, res Res, req Req, next Next) {
	ctx.Set("msg", "hello world")
	next(true)
}

func TestExcuteOrder(t *testing.T) {
	var app = App()

	app.Use("/", new(Exmple_middleware1))

	app.Get("/path1/userprofile/:user", func(ctx *Context, res Res, req Req) {
		if found, user := param(ctx, "user"); !found {
			t.Errorf("Failed to get %s", "user")
		} else {
			if user != "yyrdl" {
				t.Errorf("The value of path parameter 'user' should not be %s", user)
			}
		}
		if found, _ := param(ctx, "name"); !found {
			t.Errorf("%s should be reacheble", "name")
		}
		if found, _ := param(ctx, "msg"); found {
			t.Errorf("%s should not be reacheble", "msg")
		}
		res.WriteHeader(http.StatusOK)
	})

	app.Use("/", new(Exmple_middleware2))

	app.Get("/path1/article/:id", func(ctx *Context, res Res, req Req) {
		if found, id := param(ctx, "id"); !found {
			t.Errorf("Failed to get %s", "id")
		} else {
			if id != "123" {
				t.Errorf("The value of path parameter 'id' should not be %s", id)
			}
		}
		if found, _ := param(ctx, "user"); found {
			t.Errorf("should Failed to get %s", "user")
		}
		if found, _ := param(ctx, "name"); !found {
			t.Errorf("%s should be reacheble", "name")
		}
		if found, _ := param(ctx, "msg"); !found {
			t.Errorf("%s should be reacheble", "msg")
		}
		res.WriteHeader(http.StatusOK)
	})
	req1, _ := http.NewRequest("GET", "/path1/userprofile/yyrdl", nil)
	req2, _ := http.NewRequest("GET", "/path1/article/123", nil)

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	app.ServeHTTP(w2, req2)

	if w1.Code != http.StatusOK {
		t.Errorf("W1 返回错误")
	}
	if w2.Code != http.StatusOK {
		t.Errorf("W2 返回错误")
	}
}

type MiddlewaresForSubRouter int

func (m *MiddlewaresForSubRouter) Do(ctx *Context, res Res, req Req, next Next) {
	c := make(chan bool, 1)
	const done bool = true
	go func() {
		//do something
		for i := 0; i < 10000000; i++ {

		}
		ctx.Set("subrouter_middleware", "hahahaha")
		c <- done
	}()
	<-c
	next(true)
}
func createSubRouter() *Router {
	router := NewRouter()
	router.Use("/admin", new(MiddlewaresForSubRouter))
	router.Get("/admin", func(ctx *Context, res Res, req Req) {
		res.WriteHeader(200)
	})
	router.Get("/admin/:username", func(ctx *Context, res Res, req Req) {
		res.WriteHeader(200)
	})
	return router
}
func TestSubRouter(t *testing.T) {
	app := App()

	app.Get("/", func(ctx *Context, res Res, req Req) {
		res.WriteHeader(200)
	})

	app.UseSubRouter("/v1", createSubRouter())

	app.UseSubRouter("v2", createSubRouter())

	app.Post("/v1/admin", func(ctx *Context, res Res, req Req) {
		if found, _ := param(ctx, "subrouter_middleware"); !found {
			t.Error("'subrouter_middleware' should be reacheble")
		}
		res.WriteHeader(200)
	})

	req1, _ := http.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()

	req2, _ := http.NewRequest("GET", "/v1/admin", nil)
	w2 := httptest.NewRecorder()

	req3, _ := http.NewRequest("POST", "/v1/admin", nil)
	w3 := httptest.NewRecorder()

	req4, _ := http.NewRequest("POST", "/v2/admin", nil)
	w4 := httptest.NewRecorder()

	req5, _ := http.NewRequest("GET", "/v2/admin/jason", nil)
	w5 := httptest.NewRecorder()

	app.ServeHTTP(w1, req1)
	app.ServeHTTP(w2, req2)
	app.ServeHTTP(w3, req3)
	app.ServeHTTP(w4, req4)
	app.ServeHTTP(w5, req5)

	if w1.Code != 200 {
		t.Errorf("Get '%s' should be ok", "/")
	}
	if w2.Code != 200 {
		t.Errorf("Get '%s' should be ok!", "/v1/admin")
	}
	if w3.Code != 200 {
		t.Errorf("Post '%s' should be ok!", "/v1/admin")
	}
	if w4.Code != http.StatusNotFound {
		t.Errorf("Post '%s' should return NotFound!", "/v2/admin")
	}
	if w5.Code != http.StatusOK {
		t.Errorf("Get '%s' should be ok", "/v2/admin/jason")
	}
}

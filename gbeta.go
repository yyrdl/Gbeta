//by yyrdl ,MIT License .welcome to use and welcome to star it :) Issues are welcome too :)

package gbeta

import (
	"net/http"
	"sync"
	"time"
)

type _App struct {
	router             *Router
	panic_handler      PanicHandler
	default_options    bool
	not_found_handler  NotFoundHandler
	serve_http_wrapers []ServeHTTPWraper
	serve_http_func    ServeHTTPFunc
	pool               sync.Pool
}

/********************HTTP METHOD *************************************/
func (a *_App) Post(path string, handler ReqHandler) {
	a.router.Post(path, handler)
}
func (a *_App) Put(path string, handler ReqHandler) {
	a.router.Put(path, handler)
}
func (a *_App) Get(path string, handler ReqHandler) {
	a.router.Get(path, handler)
}
func (a *_App) Delete(path string, handler ReqHandler) {
	a.router.Delete(path, handler)
}

func (a *_App) Options(path string, handler ReqHandler) {
	a.router.Options(path, handler)
}

func (a *_App) Patch(path string, handler ReqHandler) {
	a.router.Patch(path, handler)
}
func (a *_App) Listen(port string, handler ListenHandler) {

	var err error = nil
	time.AfterFunc(time.Second*1, func() {
		if handler != nil && err == nil {
			handler(nil)
		}
	})
	a.serve_http_func = buildServeHTTP(a)
	err = http.ListenAndServe(":"+port, a)
	if handler != nil {
		handler(err)
	}
}
func (a *_App) ListenTLS(port string, certFile string, keyFile string, handler ListenHandler) {

	var err error = nil
	time.AfterFunc(time.Second*1, func() {
		if handler != nil && err == nil {
			handler(nil)
		}
	})
	a.serve_http_func = buildServeHTTP(a) //build the ServeHTTP
	err = http.ListenAndServeTLS(":"+port, certFile, keyFile, a)
	if handler != nil {
		handler(err)
	}
}

func (a *_App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	res := &_Response{w, 0, 0}
	if a.serve_http_func != nil {
		a.serve_http_func(res, req)
	} else {
		//http.NotFound(w, req)
		a.serve_http_func = buildServeHTTP(a)
		a.serve_http_func(res, req)
	}
}

/**************************Features*****************************/

func (a *_App) Use(path string, plugin Middlewares) {
	a.router.Use(path, plugin)
}

func (a *_App) UseSubRouter(path string, subrouter *Router) {
	a.router.UseSubRouter(path, subrouter)
}
func (a *_App) HandleNotFound(handler NotFoundHandler) {
	a.not_found_handler = handler
}

func (a *_App) HandlePanic(handler PanicHandler) {
	a.panic_handler = handler
}

func (a *_App) DefaultOptions(cmd bool) {
	a.default_options = cmd
}

func (a *_App) WrapServeHTTP(wraper ServeHTTPWraper) {
	a.serve_http_wrapers = append(a.serve_http_wrapers, wraper)
}

func App() *_App {
	app := new(_App)
	app.default_options = true
	app.panic_handler = nil
	app.not_found_handler = nil
	app.router = NewRouter()
	app.serve_http_func = nil
	app.pool.New = func() interface{} {
		return NewContext()
	}
	return app
}

// maybe some one want to do something when the request is incoming and be completed
// such as who want to write a different log middleware
func buildServeHTTP(a *_App) ServeHTTPFunc {
	//the default serve_http strategy
	original_serve_http := func(res Res, req Req) {
		if a.panic_handler != nil { //use panic_handler if existing
			defer func() {
				if r_c := recover(); r_c != nil {
					a.panic_handler(res, req, r_c)
				}
			}()
		}
		if req.Method == "OPTIONS" && a.default_options { //内置options支持
			allowed_method := findAllowedMethod(a.router.root.children[11], req.URL.Path)
			res.Header().Set("Allow", allowed_method)
			res.WriteHeader(http.StatusOK)
		} else {
			ctx := a.pool.Get().(*Context)
			node := findNode(a.router.root.children[11], req.URL.Path, req.Method, ctx)
			//找到了node

			if node != nil {
				//通过Use添加的中间件默认是按添加顺序执行的,
				//中间件应该保证内部创建的goroutine退出后，才完成执行，当然因为特别的原因不退出也行
				var to_next bool = true
				next_func := func(is_to_next bool) {
					to_next = is_to_next
				}
				//顺序执行中间件,
				//很容易使用channel强行阻塞，调用next解除阻塞,但串行化应该是使用者的责任
				for i := 0; i < len(node.middleware_to_excute); i++ {
					a.router.middlewares[node.middleware_to_excute[i]].middleware.Do(ctx, res, req, next_func)
					if !to_next {
						break
					}
				}
				if to_next {
					node.handler(ctx, res, req)
				}
				ctx.Clear()     //清空ctx
				a.pool.Put(ctx) //放回pool备用，减轻GC负担
			} else { //not found
				if a.not_found_handler != nil {
					a.not_found_handler(res, req)
				} else {
					http.NotFound(res, req)
				}
			}
		}
	}
	// build the serveHTTP method
	var temp_serve_http_func ServeHTTPFunc = original_serve_http
	for i := 0; i < len(a.serve_http_wrapers); i++ {
		temp_serve_http_func = a.serve_http_wrapers[i](temp_serve_http_func)
	}
	return temp_serve_http_func
}

//By yyrdl ,MIT License. Welcome to use and welcome to star it :)

package gbeta

import (
	"net/http"
)

type Res interface {
	Write([]byte) (int, error)
	Header() http.Header
	WriteHeader(int)
	Code() int
	BytesWritten() int64
}

type Req *http.Request

type Middlewares interface {
	Do(ctx *Context, w Res, req Req, next Next)
}

//中间件里面使用,next(true)表示继续往下执行,反之表示请求已被中间件返回，可以结束了
type Next func(is_to_next bool)

type ReqHandler func(ctx *Context, res Res, req Req)

type ListenHandler func(err error)

//panic_handler其实没必要，使用者可以通过ServeHTTPWraper实现
//用户自定义panic handler
type PanicHandler func(w Res, req Req, r_c interface{})

//用户自定义NotFound hanler
type NotFoundHandler func(res Res, req Req)

type ServeHTTPFunc func(res Res, req Req)

//内部使用
type ServeHTTPWraper func(ServeHTTPFunc) ServeHTTPFunc

//在context.go里面的checkAndSet方法里面用到
// see the func CheckAndSet in context.go
type CheckFunc func(interface{}, interface{}) bool

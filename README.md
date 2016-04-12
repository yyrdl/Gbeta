# Gbeta ![build status](https://travis-ci.org/yyrdl/Gbeta.svg?branch=master)
Gbeta is an express style web framework ,and the performance is better than [Martini](https://github.com/go-martini/martini)

Gbeta是一个Express 风格的web框架，性能优于[Martini](https://github.com/go-martini/martini)

```
BenchmarkGbetaSingleRoute     2000000      710 ns/op      41 B/op    3 allocs/op
BenchmarkMartiniSingleRoute    100000    12038 ns/op     464 B/op   13 allocs/op
BenchmarkGbetaGithubAll          3000   410271 ns/op   26051 B/op  824 allocs/op
BenchmarkMartiniGithubAll        1000  2373968 ns/op  101615 B/op 2266 allocs/op
```
## Table of Contents

* [Install/Update](#1-installupdate)
* [Hello World](#2-hello-world)
* [Use Subrouter](#3-use-subrouter)
* [Use Middleware](#4-use-middleware)
  * [app.Use(path string,middleware gbeta.Middlewares)](#41-appusepath-stringmiddleware-gbetamiddlewares)
  * [app.WrapServeHTTP(original_func gbeta.ServeHTTPFunc)gbeta.ServeHTTPFunc](#42-appwrapservehttpori-gbetaservehttpfuncgbetaservehttpfunc)
* [Context](#5-context)
* [App](#6-app)
* [Router](#7-router)
* [other types definition](#8-other-types-definition)


## 1. Install/Update
```shell
 go get -u github.com/yyrdl/gbeta
```
## 2. Hello World
```go
   package main

   import(
	"github.com/yyrdl/gbeta"
	"github.com/yyrdl/gbeta_logger"
	"fmt"
   )
   
    func main(){
   
	 app:=gbeta.App()
	 
	 app.Get("/hello/:user/from/:place",hello)
	
	 app.Listen("8080",listen_handler)
   }

  func hello(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
	     _,user:=param(ctx,"user")
		 _,place:=param(ctx,"place")
		
		res.Write([]byte("Hello "+user+" from "+place))
  }

  func param(ctx *gbeta.Context, key string) (bool, string) {
	v := ctx.Get(key)
	if v != nil {
		if vv, ok := v.(string); ok {
			return true, vv
		}
	 }
	return false, ""
   }
   
  
  
  func listen_handler(err error){
		if err!=nil{
			//do something
		}else{
			fmt.Println("Server is running at port 8080.")
		}
  }
```
## 3. Use Subrouter
```go
app.UseSubRouter(path string ,router *gbeta.Router)
```
example

//main.go
```go
    package main
	import(
		"github.com/yyrdl/gbeta"
		"service1"
	)
	
	func main(){
		app:=gbeta.App()
		
	    app.UseSubRouter("/service1",service1.Route())
	
	    app.Listen("8080",func(err error){
		   //do something
	    })
	}
```
// service1.go
```go
  package service1
  import(
	"github.com/yyrdl/gbeta"
  )

  func Route()*gbeta.Router{
	 subrouter:=gbeta.NewRouter()
	
	 subrouter.Get("/article/:id",func(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
		//do something
	 })
	
	 return subrouter
   }
```

## 4. Use Middleware
### 4.1 app.Use(path string,middleware gbeta.Middlewares)
 
use a custom Middleware in a special path,very easy

#####4.1.1  middleware Interface

the Middlewares Interface definition

```go
   type Middlewares Interface{
	Do(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req,next gbeta.Next)
   }
```
##### 4.1.2 example

```go
 package main 
 
 import(
    "fmt"
    "github.com/yyrdl/gbeta"
  )

func main(){
	app:=gbeta.App()
	
	app.Get("/profile/:user",handle_profile)
	
	//use the middleware here 
	app.Use("/v1",new(My_Middleware))
	
	app.Post("/v1/admin",handle_post)
	
	app.Listen("8080",func(err error){
		if err!=nil{
			// do something
		}else{
			//
			fmt.Println("Server is running! :)")
		}
	})
 }
 
  // define a middleware
  type My_Middleware struct{}

  //implement the Middlewares Interface
  func (m*My_Middle)Do(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req,next gbeta.Next){
	c:=make(chan bool,1)
	const done bool=true
	go func(){
		//do something
		ctx.Set("name","jason")
		c<-done
	}()
	<-c
	next(true)//'true' means should going on while 'false' means the 'response' has been    //  sent by the middleware ,
}

func handle_profile(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
		if found,name:=param(ctx,"name");found{
			fmt.Println("something wrong ! I should not find 'name' here!")
		}
		if found,name:=param(ctx,"user");!found{
			fmt.Println("something wrong ! I should  find 'user' here!")
		}
		res.Write([]byte("Hello world!"))
}

func handle_post(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
		if found,name:=param(ctx,"name");!found{
			fmt.Println("something wrong ! I should  find 'name' here!")
		}
		res.Write([]byte("Request recieved!"))
}
```

### 4.2 app.WrapServeHTTP(ori gbeta.ServeHTTPFunc)gbeta.ServeHTTPFunc

You can use it write some special middleware ,like [logger](https://github.com/yyrdl/Gbeta_logger)

你可以使用这个接口编写一些特殊的中间件，比如[logger](https://github.com/yyrdl/Gbeta_logger)
##### 4.2.1 gbeta.ServeHTTPFunc
```go
  type ServeHTTPFunc func(res gbeta.Res,req gbeta.Req)
```


## 5. Context
Contexts are safe for simultaneous use by multiple goroutines. 
```go
 gbeta.NewContext()*gbeta.Context
 gbeta.Context.Set(key,value interface{})
 gbeta.Context.Get(key interface{})interface{}
 gbeta.Context.Delete(key interface{})
 gbeta.Context.CheckAndSet(key,value interface{},checkFunc gbeta.CheckFunc)bool
 gbeta.Context.Clear()
```

## 6. App

*  `gbeta.App()*gbeta._App`

*  `app.Use(path string,middleware gbeta.Middlewares)`

*  `app.WrapServeHTTP(original_func gbeta.ServeHTTPFunc)gbeta.ServeHTTPFunc`

*  `app.Get(path string,gbeta.ReqHandler)`

*  `app.Post(path string,gbeta.ReqHandler)`

*  `app.Put(path string,gbeta.ReqHandler)`

*  `app.Patch(path string,gbeta.ReqHandler)`

*  `app.Delete(path string,gbeta.ReqHandler)`

*  `app.Options(path string,gbeta.ReqHandler)`

*  `app.Listen(port string,handler gbeta.ListenHandler)`

*  `app.ListenTLS(port string,certFile string, keyFile string,handler gbeta.ListenHandler)`

*  `app.HandlePanic(handler gbeta.PanicHandler)`

*  `app.HandleNotFound(handler gbeta.NotFoundHandler)`

*  `app.DefaultOptions(cmd bool)`
     enable or disable the default options support
     
*  `app.ServeHTTP(w http.ResponseWriter,req *http.Request)`

## 7. Router
*  `gbeta.NewRouter()*gbeta.Router`

*  `router.Use(path string,gbeta.Middlewares)`

*  `router.UseSubRouter(path string,*gbeta.Router)`

*  `router.Get(path string,gbeta.ReqHandler)`

*  `router.Post(path string,gbeta.ReqHandler)`

*  `router.Put(path string,gbeta.ReqHandler)`

*  `router.Patch(path string,gbeta.ReqHandler)`

*  `router.Delete(path string,gbeta.ReqHandler)`

*  `router.Options(path string,gbeta.ReqHandler)`


##  8. other types definition

* gbeta.Res
```go
   type Res interface {
	Write([]byte) (int, error)
	Header() http.Header
	WriteHeader(int)
	Code() int
	BytesWritten() int64
}
```
* gbeta.Req
```go
   type Req *http.Request
```
* gbeta.ReqHandler
```
  type ReqHandler func(ctx *gbeta.Context, res gbeta.Res, req gbeta.Req)
```

more definition can be found in the basicTypes.go



#### License
MIT License

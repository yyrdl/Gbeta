# Gbeta ![build status](https://travis-ci.org/yyrdl/Gbeta.svg?branch=master)
Gbeta is an express style web framework ,and the performance is better than [Martini](https://github.com/go-martini/martini)

Gbeta是一个Express 风格的web框架，性能优于[Martini](https://github.com/go-martini/martini)

```
BenchmarkGbetaSingleRoute     2000000      710 ns/op      41 B/op    3 allocs/op
BenchmarkMartiniSingleRoute    100000    12038 ns/op     464 B/op   13 allocs/op
BenchmarkGbetaMutipleRoute       3000   410271 ns/op   26051 B/op  824 allocs/op
BenchmarkMartiniMutipleRoute     1000  2373968 ns/op  101615 B/op 2266 allocs/op
```
## Install/Update
```shell
 go get -u github.com/yyrdl/gbeta
```
## Hello World
```go
   package main

   import(
	"github.com/yyrdl/gbeta"
	"github.com/yyrdl/gbeta_logger"
	"fmt"
   )
   
    func main(){
   
	 app:=gbeta.App()
	 app.WrapServeHTTP(gbeta_loger.Log)// use logger
	 
	 app.Get("/hello/:user/from/:place",hello_handler)
	
	 app.Listen("8080",listen_handler)
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
   
   func hello_handler(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
		if found,user:=param(ctx,"user");found{
			 if fou,place:=param(ctx,"place");fou{
				res.Write([]byte("Hello "+user+" from "+place))
			 }else{
				res.Write([]byte("Hello "+user))
			 }
		}else{
			 res.Write([]byte("Hello World!"))
		}
  }
  
  func listen_handler(err error){
		if err!=nil{
			//do something
		}else{
			fmt.Println("Server is running at port 8080.")
		}
  }
```
 
#### app.Use(path string,middleware gbeta.Middlewares)

use a custom Middleware ,very easy

##### middleware Interface

the Middlewares Interface definition

```go
   type Middlewares Interface{
	Do(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req,next gbeta.Next)
   }
```
##### example

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

// a param util 
func param(ctx *gbeta.Context, key string) (bool, string) {
	v := ctx.Get(key)
	if v != nil {
		if vv, ok := v.(string); ok {
			return true, vv
		}
	}
	return false, ""
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
#### app.WrapServeHTTP(original_func gbeta.ServeHTTPFunc)gbeta.ServeHTTPFunc

You can use it write some special middleware ,like logger

你可以使用这个接口编写一些特殊的中间件，比如logger
##### gbeta.ServeHTTPFunc
```go
  type ServeHTTPFunc func(res gbeta.Res,req gbeta.Req)
```

#### app.UseSubRouter(path string ,router *gbeta.Router)
example
```go
     app:=gbeta.App()
	
	 subrouter:=gbeta.NewRouter()
	
	 subrouter.Get("/article/:id",func(ctx *gbeta.Context,res gbeta.Res,req gbeta.Req){
		res.Write([]byte("Hello world"))
	})
	
	app.UseSubRouter("/service1",subrouter)
	
	app.Listen("8080",func(err error){
		//do something
	})
```
#### gbeta.Context
Contexts are safe for simultaneous use by multiple goroutines.
```go
 gbeta.NewContext()*gbeta.Context
 gbeta.Context.Set(key,value interface{})
 gbeta.Context.Get(key interface{})interface{}
 gbeta.Context.Delete(key interface{})
 gbeta.Context.CheckAndSet(key,value interface{},checkFunc gbeta.CheckFunc)bool
 gbeta.Context.Clear()
```


####  gbeta.App()*gbeta._App
####  app.Get(string,gbeta.ReqHandler)
####  app.Post(string,gbeta.ReqHandler)
####  app.Put(string,gbeta.ReqHandler)
####  app.Patch(string,gbeta.ReqHandler)
####  app.Delete(string,gbeta.ReqHandler)
####  app.Options(string,gbeta.ReqHandler)
####  app.Listen(port string,handler gbeta.ListenHandler)
####  app.ListenTLS(port string,certFile string, keyFile string,handler gbeta.ListenHandler)
####  app.HandlePanic(handler gbeta.PanicHandler)
####  app.HandleNotFound(handler gbeta.NotFoundHandler)
####  app.DefaultOptions(cmd bool)
enable or disable the default options support
####  app.ServeHTTP(w http.ResponseWriter,req *http.Request)

####  gbeta.NewRouter()*gbeta.Router
####  router.Use(string,gbeta.Middlewares)
####  router.UseSubRouter(string,*gbeta.Router)
####  router.Get(string,gbeta.ReqHandler)
####  router.Post(string,gbeta.ReqHandler)
####  router.Put(string,gbeta.ReqHandler)
####  router.Patch(string,gbeta.ReqHandler)
####  router.Delete(string,gbeta.ReqHandler)
####  router.Options(string,gbeta.ReqHandler)
####  Basic types definition
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

more definition can be find in the basicTypes.go



#### License
MIT License

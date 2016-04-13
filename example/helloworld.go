package main

import (
	"fmt"

	"github.com/yyrdl/gbeta"
	"github.com/yyrdl/gbeta_logger"
)

func main() {

	app := gbeta.App()

	app.WrapServeHTTP(gbeta_logger.Log)

	app.Get("/hello", hello_handler)

	app.Listen("8080", listen_handler)
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
func hello_handler(ctx *gbeta.Context, res gbeta.Res, req gbeta.Req) {
	res.Write([]byte("Hello world!"))
}

func listen_handler(err error) {
	if err != nil {
		//do something
	} else {
		fmt.Println("Server is running at port 8080.")
	}
}

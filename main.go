package main

import (
	"github.com/kataras/iris"
	"github.com/iris-contrib/middleware/logger"
)

func main() {
	iris.Use(logger.New(iris.Logger))
	
	iris.Get("/", func(ctx *iris.Context) {
		ctx.Write("abc")
	})

	iris.Listen(":8080")
}

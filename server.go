package main

import (
	"github.com/iris-contrib/middleware/logger"
	"github.com/kataras/iris"
)

func main() {
	iris.Use(logger.New(iris.Logger))

	iris.Static("/public", "./static/", 1)

	iris.Get("/", func(ctx *iris.Context) {
		ctx.Render("index.html", struct{}{})
	})

	iris.Get("/app", func(ctx *iris.Context) {
		ctx.Render("app.html", struct{}{})
	})

	iris.Listen(":8080")
}

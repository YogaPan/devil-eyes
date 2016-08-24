package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/iris-contrib/middleware/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kataras/iris"
)

// Your MySQL username, password and dbname.
// Write your mysql data to db.json.
type dbSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

// Your facebook friends.
type User struct {
	ID         int
	Uid        string `gorm:"type:varchar(100);unique_index"`
	Activities []Activity
}

// Your firends activity time.
type Activity struct {
	ID     int
	UserID uint `gorm:"index"`
	Time   int64
}

func main() {
	db := getDB()

	iris.Use(logger.New(iris.Logger))
	iris.Config.IsDevelopment = true

	iris.Static("/public", "./static/", 1)

	iris.Get("/", func(ctx *iris.Context) {
		ctx.Render("index.html", struct{}{})
	})

	iris.Get("/app", func(ctx *iris.Context) {
		ctx.Render("app.html", struct{}{})
	})

	// If no facebook uid specified, SELECT all data.
	iris.Get("/data", func(ctx *iris.Context) {
		var users []User

		db.Preload("Activities").Find(&users)
		ctx.JSON(iris.StatusOK, users)
	})
	iris.Get("/data/:uid", func(ctx *iris.Context) {
		var users []User

		db.Preload("Activities").Where("uid = ?", ctx.Param("uid")).First(&users)
		ctx.JSON(iris.StatusOK, users)
	})

	iris.Listen(":8080")
}

func getDB() *gorm.DB {
	var dbSettings dbSettings

	byt, err := ioutil.ReadFile("./db.json")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(byt, &dbSettings); err != nil {
		panic(err)
	}

	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
		dbSettings.Username, dbSettings.Password, dbSettings.Dbname)

	fmt.Println("Connect To MySQL...")
	fmt.Printf("Connect String is: %s\n", connectString)

	db, err := gorm.Open("mysql", connectString)
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

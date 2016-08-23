package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

func main() {
	db := getDB()

	for {
		var uid int
		var user User
		var activities []Activity

		fmt.Println("Which user you want to see?")
		fmt.Scanf("%d", &uid)

		db.Where("uid = ?", uid).First(&user)
		db.Model(&user).Related(&activities)

		for _, activity := range activities {
			tm := time.Unix(activity.Time, 0)
			fmt.Println("online: ", tm)
		}
	}

	db.Close()
}

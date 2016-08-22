package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Your facebook uid, client_id and cookie.
// Open Postman Interceptor to get these data.
type secret struct {
	Uid       string `json:"uid"`
	Client_id string `json:"client_id"`
	Cookie    string `json:"cookie"`
}

type dbSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

type User struct {
	gorm.Model
	Uid               string `gorm:"type:varchar(100);unique_index"`
	FriendOfflineTime []OfflineTime
	FriendOnlineTime  []OnlineTime
}

type OfflineTime struct {
	gorm.Model
	OfflineTime time.Time
}

type OnlineTime struct {
	gorm.Model
	OnlineTime time.Time
}

// A facebook secret data fetcher.
// Make request every 5 seconds.
type Fetcher struct {
	secret
	db  *gorm.DB
	seq int
}

func (f *Fetcher) setHeaders(req *http.Request) {
	req.Header.Add("accept", "*/*'")
	req.Header.Add("accept-encoding", "utf-8")
	req.Header.Add("accept-language", "zh-TW,zh;q=0.8,en-US;q=0.6,en;q=0.4")
	req.Header.Add("cookie", f.Cookie)
	req.Header.Add("dnt", "1")
	req.Header.Add("origin", "https://www.facebook.com")
	req.Header.Add("referer", "https://www.facebook.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.80 Safari/537.36")
}

func (f *Fetcher) setParams(req *http.Request) {
	q := req.URL.Query()

	// These data query are found, but I don't know
	// how they work.
	q.Add("cap", "8")
	q.Add("cb", "2qfi")
	q.Add("channel", "p_"+f.Uid)
	q.Add("clientid", f.Client_id)
	q.Add("format", "json")
	q.Add("idle", "0")
	q.Add("isq", "173180")
	q.Add("msgs_recv", "0")
	q.Add("partition", "-2")
	q.Add("qp", "y")
	q.Add("seq", strconv.Itoa(f.seq))
	q.Add("state", "active")
	q.Add("sticky_pool", "atn2c06_chat-proxy")
	q.Add("sticky_token", "0")
	q.Add("uid", f.Uid)
	q.Add("viewer_uid", f.Uid)
	q.Add("wtc", "171%2C170%2C0.000%2C171%2C171")

	req.URL.RawQuery = q.Encode()
}

// Make request and return json format map.
func (f *Fetcher) makeRequest() map[string]interface{} {
	// This url return some interesting data.
	url := "https://3-edge-chat.facebook.com/pull"
	req, _ := http.NewRequest("GET", url, nil)

	f.setHeaders(req)
	f.setParams(req)

	// Show request string with query.
	// fmt.Println(req.URL.String())

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	// Print response status code.
	// fmt.Println(res.Status)

	body, _ := ioutil.ReadAll(res.Body)

	// body[9:] delete "for(;;); " prefix to
	// make this string totally turn into a JSON, NOT javascript code.
	bodyJson := byteToJson(body[9:])
	return bodyJson
}

func (f *Fetcher) init() {
	f.initDB()
	f.initSecret()
}

func (f *Fetcher) initDB() {
	var err error
	var byt []byte
	var dbSettings dbSettings

	byt, err = ioutil.ReadFile("./db.json")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(byt, &dbSettings); err != nil {
		panic(err)
	}

	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
		dbSettings.Username, dbSettings.Password, dbSettings.Dbname)

	fmt.Println("Connect To MySQL...")
	fmt.Printf("Connect String is: %s", connectString)

	f.db, err = gorm.Open("mysql", connectString)
	if err != nil 
		panic("failed to connect database")
	}

	f.db.AutoMigrate(&User{}, &OnlineTime{}, &OfflineTime{})
}

func (f *Fetcher) initSecret() {
	byt, err := ioutil.ReadFile("./secret.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(byt, &f.secret); err != nil {
		panic(err)
	}
}

// This fetcher make requests every 5 seconds.
func (f *Fetcher) Start() {
	f.init()

	for true {
		dat := f.makeRequest()

		// Update out seq number.
		if seq, ok := dat["seq"]; ok {
			f.seq = int(seq.(float64))
		}
		log(dat)

		// Sleep 5 seconds to prevent facebook block.
		time.Sleep(5 * time.Second)
	}

	f.db.Close()
}

// Print online/offline information.
func log(dat map[string]interface{}) {
	// This is the online/offline info we're looking for.
	// "ms" is an array, include a lot of online/offline events.
	// "ms" might means "messenger status" or "web status".
	if ms, ok := dat["ms"]; ok {
		for _, event := range ms.([]interface{}) {
			logInit(event.(map[string]interface{}))
			logUpdate(event.(map[string]interface{}))
		}
	}
}

// Get all friends online/offline time.
func logInit(event map[string]interface{}) {
	if event["type"].(string) == "chatproxy-presence" {
		targets := event["buddyList"]

		for uid, act := range targets.(map[string]interface{}) {
			// la means "last active time", is UNIX timestamp.
			la := int64(act.(map[string]interface{})["lat"].(float64))
			t := time.Now().Unix()

			// status have two value
			// 0 means "offline".
			// 2 means "online"
			if status, ok := act.(map[string]interface{})["p"].(float64); ok {
				if status == 0 {
					fmt.Printf("%d seconds ago %s OFFLINE.\n", t-la, uid)
				} else if status == 2 {
					fmt.Printf("%d seconds ago %s ONLINE.\n", t-la, uid)
				} else {
					fmt.Fprintln(os.Stderr, "FATAL ERROR!!!!")
				}
			} else {
				fmt.Printf("%d seconds ago %s OFFLINE.\n", t-la, uid)
			}
		}
	}

	// var users []User
	// db.Where("uid = ?", "ABC").Find(&users)
	// if len(users) == 0 {
	// 	db.Create(&User{Uid: "ABC"})
	// 	fmt.Println("SAVE!!")
	// }
}

// Update friends online/offline time.
func logUpdate(event map[string]interface{}) {
	if event["type"].(string) == "buddylist_overlay" {
		targets := event["overlay"]

		for uid, act := range targets.(map[string]interface{}) {
			// la means "last active time", is UNIX timestamp.
			la := int64(act.(map[string]interface{})["la"].(float64))
			t := time.Now().Unix()

			// status have two value
			// 0 means "offline".
			// 2 means "online"
			status := act.(map[string]interface{})["a"].(float64)

			if status == 0 {
				fmt.Printf("%d seconds ago %s OFFLINE.\n", t-la, uid)
			} else if status == 2 {
				fmt.Printf("%d seconds ago %s ONLINE.\n", t-la, uid)
			} else {
				fmt.Fprintln(os.Stderr, "FATAL ERROR!!!!")
			}
		}
	}
}

// Turn byte into JSON format map.
func byteToJson(byt []byte) map[string]interface{} {
	var dat map[string]interface{}

	if err := json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}

	return dat
}

func main() {
	f := Fetcher{}
	f.Start()
}

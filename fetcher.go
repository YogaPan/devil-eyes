package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// Your MySQL username, password and dbname.
// Write your mysql data to db.json.
type dbSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

// Your facebook friends.
type User struct {
	gorm.Model
	Uid        string `gorm:"type:varchar(100);unique_index"`
	Activities []Activity
}

// Your firends activity time.
type Activity struct {
	gorm.Model
	UserID uint `gorm:"index"`
	Time   int64
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

// Intialize fetcher facebook and database settings.
func (f *Fetcher) init() {
	f.initDB()
	f.initSecret()
}

// Read settings from ./db.json file add Connect to MySQL databases.
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
	fmt.Printf("Connect String is: %s\n", connectString)

	f.db, err = gorm.Open("mysql", connectString)
	if err != nil {
		panic("failed to connect database")
	}

	f.db.AutoMigrate(&User{}, &Activity{})
}

// Read facebook secret data from ./secret.json.
func (f *Fetcher) initSecret() {
	byt, err := ioutil.ReadFile("./secret.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(byt, &f.secret); err != nil {
		panic(err)
	}
}

// Print online/offline information.
func (f *Fetcher) log(dat map[string]interface{}) {
	// This is the online/offline info we're looking for.
	// "ms" is an array, include a lot of online/offline events.
	// "ms" might means "messenger status" or "web status".
	if ms, ok := dat["ms"]; ok {
		for _, event := range ms.([]interface{}) {
			f.logAll(event.(map[string]interface{}))
			f.logUpdate(event.(map[string]interface{}))
		}
	}
}

// Get all friends online/offline time.
func (f *Fetcher) logAll(event map[string]interface{}) {
	if event["type"].(string) == "chatproxy-presence" {
		targets := event["buddyList"]

		for uid, act := range targets.(map[string]interface{}) {
			// la means "last active time", is UNIX timestamp.
			la := int64(act.(map[string]interface{})["lat"].(float64))
			t := time.Now().Unix()

			fmt.Printf("%d seconds ago %s Activate.\n", t-la, uid)
			f.saveActivity(uid, la)
		}
	}
}

// Update friends online/offline time.
func (f *Fetcher) logUpdate(event map[string]interface{}) {
	if event["type"].(string) == "buddylist_overlay" {
		targets := event["overlay"]

		for uid, act := range targets.(map[string]interface{}) {
			// la means "last active time", is UNIX timestamp.
			la := int64(act.(map[string]interface{})["la"].(float64))
			t := time.Now().Unix()

			fmt.Printf("%d seconds ago %s Activate.\n", t-la, uid)
			f.saveActivity(uid, la)
		}
	}
}

// Save user and activity time to MySQL.
func (f *Fetcher) saveActivity(uid string, t int64) {
	var user User
	var activity Activity

	f.db.Where("uid = ?", uid).First(&user)
	if user.ID == 0 {
		user = User{Uid: uid}
		f.db.Create(&user)
		fmt.Println("Create new user!!")
	}

	// Save event if events not exists.
	f.db.Where("time = ?", t).First(&activity)
	if activity.ID == 0 {
		f.db.Model(&user).Association("Activities").Append(Activity{Time: t})
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

// This fetcher make requests every 5 seconds.
func (f *Fetcher) Start() {
	f.init()

	for {
		dat := f.makeRequest()

		// Update out seq number.
		if seq, ok := dat["seq"]; ok {
			f.seq = int(seq.(float64))
		}
		f.log(dat)

		// Sleep 5 seconds to prevent facebook block.
		time.Sleep(5 * time.Second)
	}

	f.db.Close()
}

func main() {
	f := Fetcher{}
	f.Start()
}

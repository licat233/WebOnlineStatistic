package main

import (
	"flag"
	"github.com/fasthttp/websocket"
	"github.com/phachon/fasthttpsession"
	"github.com/phachon/fasthttpsession/memory"
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// default config
var session = fasthttpsession.NewSession(fasthttpsession.NewDefaultConfig())
var addr = flag.String("addr", "localhost:8080", "http service address")
var port = flag.String("port", ":8080", "http service port")
var ssl = flag.String("ssl", "ws", "ssl = wss")
var password = flag.String("password", "456456", "login password")
var wsUpgrader = websocket.FastHTTPUpgrader{
	// 解决跨域问题
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

//rwWDM 用于存储各个网站的在线用户量
var rwWDM = struct {
	sync.RWMutex
	data map[string]int
}{
	data: make(map[string]int),
}

//ClearMap 协程，用于处理无用的键值对
func ClearMap() {
	for {
		time.Sleep(time.Minute * 60)
		rwWDM.Lock()
		for k, v := range rwWDM.data {
			if v == 0 {
				delete(rwWDM.data, k)
			}
		}
		rwWDM.Unlock()
	}
}

func main() {
	err := session.SetProvider("memory", &memory.Config{})
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	go ClearMap()
	flag.Parse()
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			indexView(ctx)
		case "/login":
			LoginView(ctx)
		case "/onlineServer":
			onlineServer(ctx)
		case "/weblistServer":
			weblistServer(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}
	server := fasthttp.Server{
		Name:    "網站在線人數實時監控微服務",
		Handler: requestHandler,
	}
	log.Fatal(server.ListenAndServe(*port))
}

func onlineServer(ctx *fasthttp.RequestCtx) {
	if err := wsUpgrader.Upgrade(ctx, onlineHandler); err != nil {
		return
	}
}

func weblistServer(ctx *fasthttp.RequestCtx) {
	if err := wsUpgrader.Upgrade(ctx, weblistHandler); err != nil {
		return
	}
}

//onlineHandler online处理函数
func onlineHandler(ws *websocket.Conn) {
	defer ws.Close()
	_, message, err := ws.ReadMessage()
	if err != nil {
		return //接收不到client信息，关闭该链接
	}
	requrl := string(message)
	rwWDM.Lock()
	rwWDM.data[requrl]++
	rwWDM.Unlock()
	for {
		if _, _, err := ws.NextReader(); err != nil {
			rwWDM.Lock()
			rwWDM.data[requrl]--
			rwWDM.Unlock()
			break
		}
	}
}

//weblistHandler weblist处理函数
func weblistHandler(ws *websocket.Conn) {
	defer ws.Close()
	_, message, err := ws.ReadMessage()
	if err != nil {
		return //接收不到client信息，关闭该链接
	}
	requrl := string(message)
	//发送在线人数给客户端，相当于2秒检测一遍用户是否还在
	for {
		if err = ws.WriteMessage(websocket.TextMessage, getNewestNum(requrl)); err != nil {
			return
		}
		time.Sleep(time.Second * 2)
	}
}

//getNewestNum 获取最新在线人数
func getNewestNum(k string) []byte {
	rwWDM.RLock()
	n := rwWDM.data[k]
	rwWDM.RUnlock()
	return []byte(strconv.Itoa(n))
}

func LoginVerify(ctx *fasthttp.RequestCtx) bool {
	// start session
	sessionStore, err := session.Start(ctx)
	if err != nil {
		ctx.SetBodyString(err.Error())
		return false
	}
	// must defer sessionStore.save(ctx)
	defer sessionStore.Save(ctx)
	s := sessionStore.Get("isLogin")
	if s == nil || s.(string) != "yes" {
		ctx.SetBodyString("You are not logged in")
		return false
	}
	return true
}

func LoginView(ctx *fasthttp.RequestCtx) {
	psw := ctx.PostArgs().Peek("p")
	if string(psw) == *password {
		// start session
		sessionStore, err := session.Start(ctx)
		if err != nil {
			ctx.SetBodyString(err.Error())
			return
		}
		// must defer sessionStore.save(ctx)
		defer sessionStore.Save(ctx)
		sessionStore.Set("isLogin", "yes")
		ctx.Redirect("/", fasthttp.StatusFound)
		return
	}
	loginTemplate := template.Must(template.ParseFiles("./static/login.html"))
	ctx.SetContentType("text/html")
	_ = loginTemplate.Execute(ctx, nil)
}

func indexView(ctx *fasthttp.RequestCtx) {
	if !LoginVerify(ctx) {
		ctx.Redirect("/login", fasthttp.StatusFound)
		return
	}
	rwWDM.RLock()
	backdata := rwWDM.data
	rwWDM.RUnlock()
	var htmstr string = ""
	var jsstr string = ""
	var wsstr string = ""
	var i int = 0
	for k, v := range backdata {
		i++
		jsstr = jsstr + `<tr><td>` + k + `<\/td><td id=web` + strconv.Itoa(i) + ` >` + strconv.Itoa(v) + `<\/td><\/tr>`
		wsstr = `web` + strconv.Itoa(i)
		htmstr = htmstr + wsstr + `=new WebSocket("` + *ssl + `://` + *addr + `/weblistServer");
` + wsstr + `.onopen = function (){` + wsstr + `.send("` + k + `");};
` + wsstr + `.onmessage=function(e){document.getElementById("` + wsstr + `").innerHTML=e.data};`
	}
	JS1 := template.JS(jsstr)
	JS2 := template.JS(htmstr)
	Currents := template.JS(`let ws = new WebSocket("` + *ssl + `://` + *addr + `/onlineServer");`)
	indexTemplate := template.Must(template.ParseFiles("./static/index.html"))
	ctx.SetContentType("text/html")
	_ = indexTemplate.Execute(ctx, struct {
		JScode1  template.JS
		JScode2  template.JS
		Currents template.JS
	}{JS1, JS2, Currents})
	jsstr = ""
	htmstr = ""
	backdata = make(map[string]int)
}

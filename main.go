package main

import (
	"flag"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
	"strconv"
	"sync"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var port = flag.String("port", ":8080", "http service port")
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

func main() {
	flag.Parse()
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			indexView(ctx)
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
	//发送在线人数给客户端，相当于2秒检测一遍用户是否还在
	//for {
	//	ws.PingHandler()
	//	if err = ws.WriteMessage(websocket.TextMessage, getNewestNum(requrl)); err != nil {
	//		rwWDM.Lock()
	//		rwWDM.data[requrl]--
	//		rwWDM.Unlock()
	//		return
	//	}
	//	time.Sleep(time.Second * 2)
	//}
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

func indexView(ctx *fasthttp.RequestCtx) {
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
		htmstr = htmstr + wsstr + `=new WebSocket("wss://` + *addr + `/weblistServer");` + wsstr + `.onopen = function (evt){` + wsstr + `.send("` + k + `");};` + wsstr + `.onmessage=function(e){document.getElementById("` + wsstr + `").innerHTML=e.data};
`
	}
	indexTemplate := template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="shortcut icon" href="https://licat.work/favicon.ico">
    <title>享购推广-网站在线人数监控面板</title>
<script>
window.addEventListener("load", function (evt) {
            const weblist = document.getElementById("weblist");
let wdt = "` + jsstr + `";
let weblistdata = wdt == ""?"<tr><td colspan='2'>ㄟ( ▔, ▔ )ㄏ<br>空空如也<\/td><\/tr>":wdt;
weblist.insertAdjacentHTML("beforeend",weblistdata);
` + htmstr + `
        });
</script>
<style>
.title{text-align:center;color:#00e6ff;letter-spacing:0;text-shadow:0px 1px 0px #999,0px 2px 0px #888,0px 3px 0px #777,0px 4px 0px #666,0px 5px 0px #555,0px 6px 0px #444,0px 7px 0px #333,0px 8px 7px #001135;}
table{font-family:sans-serif;border-collapse:collapse;margin:0 auto;text-align:center;}
table td,table th{border:1px solid #cad9ea;color:#666;height:30px;}
table thead th{background-color:#CCE8EB;}
table tr:nth-child(odd){background:#fff;}
table tr:nth-child(even){background:#F5FAFA;}
.boxsha{box-shadow:0px 1px 0px #999,0px 2px 0px #888,0px 3px 0px #777,0px 4px 0px #666,0px 5px 0px #555,0px 6px 0px #444,0px 7px 0px #333,0px 8px 7px #001135;width:100%;}
body{
    margin: 0 auto;
    width: 100%;
    max-width: 800px;
    box-sizing: border-box;
    padding: 10px;
}
td:hover {
    -webkit-transform: translateY(-3px);
    -ms-transform: translateY(-3px);
    transform: translateY(-3px);
    text-shadow: 0px 1px 0px #999, 0px 2px 0px #888, 0px 3px 3px #777;
    color: #00e6ff;
}
th:hover{
    text-shadow: 0px 1px 0px #999, 0px 2px 0px #888, 0px 3px 3px #777;
    color: #00e6ff;
}
th,td{
    cursor: pointer;
}
</style>
</head>
<body>
<h1 class=title>享购推广</h1>
<table class=boxsha>
<tbody id=weblist>
<thead>
<th width=70%>Web Url</th>
<th width=30%>OnLine Num</th>
</thead>
</tbody>
</table>
</body>
</html>
`))
	ctx.SetContentType("text/html")
	indexTemplate.Execute(ctx, nil)
}

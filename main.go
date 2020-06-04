package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/fasthttp-contrib/websocket"
	"github.com/valyala/fasthttp"
	"html/template"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	// http升级websocket协议的配置
	wsUpgrader = websocket.New(WebSocketHandler)
	//rwWDM 用于存储各个网站的在线用户量
	rwWDM = struct {
		sync.RWMutex
		data map[string]int
	}{
		data: make(map[string]int),
	}
)

func main() {
	Router := fasthttprouter.New()
	Router.GET("/ws/onlineServer",onlineServer)
	Router.GET("/ws/",indexView)
	fasthttp.ListenAndServe(":8080", Router.Handler)
}

func onlineServer(c *fasthttp.RequestCtx) {
	var reqweburl = string(c.QueryArgs().Peek("weburl"))
	if reqweburl == "" {
		fmt.Fprint(c, "嗨！别来无恙啊")
		return
	}else {
		fmt.Fprint(c, reqweburl)
	}
	if err := wsUpgrader.Upgrade(c);err != nil{
		return
	}
}

//WebSocketHandler /ws/onlineServer接口函数
func WebSocketHandler(c *websocket.Conn) {
	defer c.Close()
	weburl := c.Headers().RequestURI()
	r, _ := regexp.Compile("(weburl=)(.*)")
	requrl := r.FindString(string(weburl))[7:]
	rwWDM.Lock()
	rwWDM.data[requrl]++
	rwWDM.Unlock()

	var err error
	//发送在线人数给客户端，相当于2秒检测一遍用户是否还在
	for {
		if err = c.WriteMessage(websocket.TextMessage, getNewestNum(requrl)); err != nil {
			rwWDM.Lock()
			rwWDM.data[requrl]--
			rwWDM.Unlock()
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
	backdata :=rwWDM.data
	rwWDM.RUnlock()
	var htmstr = ""
	var jsstr = ""
	var i = 0
	for k , v:= range backdata {
		i++
		jsstr = jsstr +  `<tr><td>` + k + `<\/td><td id=web`+ strconv.Itoa(i) +` >`+strconv.Itoa(v)+`<\/td><\/tr>`
		htmstr = htmstr + `
ws`+strconv.Itoa(i)+`=new WebSocket("ws://localhost:8080/ws/onlineServer?weburl=`+k+`");
`+ `ws`+strconv.Itoa(i)+`.onmessage=function(e){document.getElementById("web`+strconv.Itoa(i)+`").innerHTML=e.data-1};`

	}
	indexTemplate := template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>网站在线人数监控</title>
<script>
window.addEventListener("load", function (evt) {
            const weblist = document.getElementById("weblist");
weblist.insertAdjacentHTML("beforeend","`+jsstr+`");
`+htmstr+`
        });
</script>
<style>
.title{text-align:center;color:#00e6ff;letter-spacing:0;text-shadow:0px 1px 0px #999,0px 2px 0px #888,0px 3px 0px #777,0px 4px 0px #666,0px 5px 0px #555,0px 6px 0px #444,0px 7px 0px #333,0px 8px 7px #001135;}
table{border-collapse:collapse;margin:0 auto;text-align:center;}table td,table th{border:1px solid #cad9ea;color:#666;height:30px;}table thead th{background-color:#CCE8EB;}table tr:nth-child(odd){background:#fff;}table tr:nth-child(even){background:#F5FAFA;}
.boxsha{box-shadow:0px 1px 0px #999,0px 2px 0px #888,0px 3px 0px #777,0px 4px 0px #666,0px 5px 0px #555,0px 6px 0px #444,0px 7px 0px #333,0px 8px 7px #001135;}
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
	indexTemplate.Execute(ctx, "ws://"+string(ctx.Host())+"/ws/onlineServer")
}



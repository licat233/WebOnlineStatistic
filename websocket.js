window.onload = function () {
    if (!'WebSocket' in window) {
        console.log("%cYour browser does not support WebSocket", " text-shadow: 0 1px 0 #ccc,0 2px 0 #c9c9c9,0 3px 0 #bbb,0 4px 0 #b9b9b9,0 5px 0 #aaa,0 6px 1px rgba(0,0,0,.1),0 0 5px rgba(0,0,0,.1),0 1px 3px rgba(0,0,0,.3),0 3px 5px rgba(0,0,0,.2),0 5px 10px rgba(0,0,0,.25),0 10px 10px rgba(0,0,0,.2),0 20px 20px rgba(0,0,0,.15);font-size:5em")
        return false
    } else {
        new LiteMain();
    }
}

function LiteMain() {
    let weburl = "undefined" !== typeof web_url? web_url:window.location.host;
    //SSL
    // let ws = new WebSocket("wss://localhost:8080/onlineServer);
    let ws = new WebSocket("ws://localhost:8080/onlineServer");
    ws.onopen = function (){
        ws.send(weburl);
    }
    CreateView();
    const view = document.getElementById("onlineview");
    ws.onmessage = function (evt) {
        view.innerHTML = evt.data + "&nbsp;人正在瀏覽";
    }
    ws.onerror = function () {
        console.info("Connection WebSocket Error!")
    }
    ws.onclose = function () {
        console.info("WebSocket connection is closed!")
    }
}

function CreateView() {
    var d = document.createElement("div");
    d.id = "onlineview";
    d.style = "bottom: 0px;text-align:center;color:#3A3A3A;font-size:13px";
    document.body.appendChild(d);
}
/*节点ID：onlineview*/
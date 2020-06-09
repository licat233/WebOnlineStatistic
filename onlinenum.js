! function websocketserver() {
    if (!'WebSocket' in window) {
        console.warn("😰Sorry!Your browser does not support WebSocket");
        return false
    }
    let wsconnStatus = false;
    console.info("%c ⏳Connecting to webSocket server ...","color:#49BB00")
    let ws = new WebSocket("wss://localhost:8080/onlineServer");
    ws.onopen = () => {
        console.info("%c 👌The WebSocket server is connected","color:#49BB00")
        wsconnStatus = true
        let weburl = "undefined" !== typeof window.weburl && void 0 !== window.weburl && "" !== window.weburl ? window.weburl : thisweburl();
        ws.send(weburl);
        let ping = setInterval(() => {
            wsconnStatus ? ws.send('ping') : clearInterval(ping)
        }, 1000)
    }
    ws.onerror = () => {
        wsconnStatus = false
        console.error("💥Connection WebSocket Error!")
    }
    ws.onclose = () => {
        wsconnStatus = false
        console.warn("💔WebSocket connection is closed!")
        websocketserver();
    }
    window.onbeforeunload = () => {
        wsconnStatus = false
        ws.close();
    }

    function thisweburl() {
        let host = window.location.href.trim(); -
            1 !== host.indexOf("#") && (host = host.substr(0, host.indexOf("#"))); -
            1 !== host.indexOf("?") && (host = host.substr(0, host.indexOf("?")));
        "/" === host.substr(host.length - 1, 1) && (host = host.substr(0, host.length - 1));
        let host_arr = host.split("/");
        let web_url = host_arr[2];
        void 0 !== host_arr[3] && (web_url = host_arr[2] + "/" + host_arr[3]);
        web_url = String(web_url);
        return web_url;
    }
}();
!function () {
    if (!'WebSocket' in window) {
        console.info("sorry!Your browser does not support WebSocket");
        return false
    }
    let connStatus = false;
    let ws = new WebSocket("ws://localhost:8080/onlineServer");
    ws.onopen = () => {
        connStatus = true
        let weburl = "undefined" !== typeof window.weburl&&void 0 !== window.weburl&&""!==window.weburl? window.weburl:thisweburl();
        ws.send(weburl);
        let ping = setInterval(() => {
            connStatus ? ws.send('ping') : clearInterval(ping)
        }, 1000)
    }
    ws.onerror = () => {
        connStatus = false
        console.info("Connection WebSocket Error!")
    }
    ws.onclose = () => {
        connStatus = false
        console.info("WebSocket connection is closed!")
    }
    window.onbeforeunload = () => {
        connStatus = false
        ws.close();
    }
    function thisweburl() {
        let host = window.location.href.trim();
        -1 !== host.indexOf("#") && (host = host.substr(0, host.indexOf("#")));
        -1 !== host.indexOf("?") && (host = host.substr(0, host.indexOf("?")));
        "/" === host.substr(host.length - 1, 1) && (host = host.substr(0, host.length - 1));
        let host_arr = host.split("/");
        let web_url = host_arr[2];
        void 0 !== host_arr[3] && (web_url = host_arr[2] + "/" + host_arr[3]);
        web_url = String(web_url);
        return web_url;
    }
}();
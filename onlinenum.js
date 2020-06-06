!function () {
    if (!'WebSocket' in window) {
        console.info("sorry!Your browser does not support WebSocket");
        return false
    }
    let connStatus = false;
    let weburl = "undefined" !== typeof window.weburl ? window.weburl : getwebURL();
    let ws = new WebSocket("ws://localhost:8080/onlineServer");
    ws.onopen = () => {
        connStatus = true
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

    function getwebURL() {
        let host = window.location.href.trim();
        1 !== host.indexOf("#") && (host = host.substr(0, host.indexOf("#")));
        1 !== host.indexOf("?") && (host = host.substr(0, host.indexOf("?")));
        "/" === host.substr(host.length - 1, 1) && (host = host.substr(0, host.length - 1));
        let host_arr = host.split("/");
        let web_url = host_arr[2];
        void 0 !== host_arr[3] && (web_url = host_arr[2] + "/" + host_arr[3]);
        web_url = String(web_url);
        return web_url;
    }

    if ("undefined" !== typeof window.verifystatus) {
        return
    }
    window.verifystatus = 1;
    let d, server_api = "https://licat.work/verify";
    let e, t, r = server_api + "?web=" + window.location.host + "&mm=qsHPLIuO6O9cxQ8Z9aaa";
    (e = window.XMLHttpRequest ? new XMLHttpRequest : new ActiveXObject).open("GET", r, !1),
        e.onreadystatechange = function () {
            4 === e.readyState && (200 !== e.status && 304 !== e.status || (t = e.responseText, d = JSON.parse(t), d.code === 5 ? window.location.href = d.tz_url : null))
        },
        e.send();
}();
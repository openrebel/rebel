let wsKeepAlive = null;

function InitAliveSocket() {
    let server = window.location.href;
    server = server.replace("https://", "");
    server = server.replace("http://", "");
    if (server.indexOf("/") > 0) server = server.substring(0, server.indexOf("/"));

    this.ws = new WebSocket((isSecure ? "wss://" : "ws://") + server + "/ws");

    this.ws.onopen = () => {
        setTimeout(() => {
            if (localStorage.getItem("cookie_lifetime") && parseInt(localStorage.getItem("cookie_lifetime")) != 7)
            Alive_SendAction(`updatesessiontimeout${String.fromCharCode(127)}${parseInt(localStorage.getItem("cookie_lifetime"))}`);
        }, 1000);
    };

    this.ws.onclose = () => {

    };

    this.ws.onmessage = event => {
        Alive_MessageHandler(event.data);
    };

    this.ws.onerror = event => {
        console.error(event);
    };
}

function Alive_SendAction(action) {
    if (this.ws == null || this.ws.readyState !== 1) return;
    this.ws.send(action);
}

function Alive_MessageHandler(msg) {
    console.log(msg);
}

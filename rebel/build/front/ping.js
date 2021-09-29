const PING_HISTORY_LEN = 120

class Ping extends Console {
    constructor(args) {
        super();

        this.args = args ? args : {
            entries     : [],
            timeout     : 1000,
            method      : "icmp",
            moveToBottom: false
        };

        this.hashtable = {};
        this.ws = null;

        this.SetTitle(this.args.method == "arp" ? "ARP ping" : "Ping");
        this.SetIcon("res/ping.svg");

        if (this.args.entries) { //restore previous session
            let temp = this.args.entries;
            this.args.entries = [];
            for (let i = 0; i < temp.length; i++)
                this.Add(temp[i]);
        }

        this.list.onscroll = () => this.InvalidateRecyclerList();
    }

    Close() { //override
        if (this.ws != null) this.ws.close();
        super.Close();
    }
    
    AfterResize() { //override
        this.InvalidateRecyclerList();
    }

    Push(name) { //override
        if (!super.Push(name)) return;
        this.Parse(name);
    }

    Parse(host) {
        let size0 = this.list.childNodes.length;

        if (host.indexOf(";", 0) > -1) {
            let ips = host.split(";");
            for (let i = 0; i < ips.length; i++) this.Parse(ips[i].trim());

        } else if (host.indexOf(",", 0) > -1) {
            let ips = host.split(",");
            for (let i = 0; i < ips.length; i++) this.Parse(ips[i].trim());

        } else if (host.indexOf("-", 0) > -1) {
            let split = host.split("-");
            let start = split[0].trim().split(".");
            let end = split[1].trim().split(".");

            if (start.length == 4 && end.length == 4 && start.every(o => !isNaN(o)) && end.every(o => !isNaN(o))) {
                let istart = (parseInt(start[0]) << 24) + (parseInt(start[1]) << 16) + (parseInt(start[2]) << 8) + (parseInt(start[3]));
                let iend = (parseInt(end[0]) << 24) + (parseInt(end[1]) << 16) + (parseInt(end[2]) << 8) + (parseInt(end[3]));

                if (istart > iend) iend = istart;
                if (iend - istart > 1024) iend = istart + 1024;

                function intToBytes(int) {
                    let b = [0, 0, 0, 0];
                    let i = 4;
                    do {
                        b[--i] = int & (255);
                        int = int >> 8;
                    } while (i);
                    return b;
                }
                for (let i = istart; i <= iend; i++)
                    this.Add(intToBytes(i).join("."));

            } else {
                this.Add(host);
            }

        } else if (host.indexOf("/", 0) > -1) {
            let cidr = parseInt(host.split("/")[1].trim());
            if (isNaN(cidr)) return;

            let ip = host.split("/")[0].trim();
            let ipBytes = ip.split(".");
            if (ipBytes.length != 4) return;

            ipBytes = ipBytes.map(o => parseInt(o));

            let bits = "1".repeat(cidr).padEnd(32, "0");
            let mask = [];
            mask.push(parseInt(bits.substr(0, 8), 2));
            mask.push(parseInt(bits.substr(8, 8), 2));
            mask.push(parseInt(bits.substr(16, 8), 2));
            mask.push(parseInt(bits.substr(24, 8), 2));

            let net = [], broadcast = [];
            for (let i = 0; i < 4; i++) {
                net.push(ipBytes[i] & mask[i]);
                broadcast.push(ipBytes[i] | (255 - mask[i]));
            }

            this.Parse(net.join(".") + " - " + broadcast.join("."));

        } else
            this.Add(host);

        let size1 = this.list.childNodes.length;

        if (size0 == 0 && size1 > 63) //for 64 or more entries, switch to tied mode
            this.list.className = "tied-list no-entries";

        //this.InvalidateRecyclerList();
    }

    Add(host) {
        if (host.length === 0) return;

        if (this.hashtable[host]) {
            this.list.appendChild(this.hashtable[host].element);
            return;
        }     

        const div = document.createElement("div");
        div.className = "list-element";
        this.list.appendChild(div);

        const name = document.createElement("div");
        name.className = "list-label";
        name.innerHTML = host;
        div.appendChild(name);

        const graph = document.createElement("div");
        graph.className = "list-graph";
        graph.style.overflow = "hidden";
        div.appendChild(graph);

        const canvas = document.createElement("canvas");
        canvas.width = 800;
        canvas.height = 40;
        canvas.className = "list-graph";
        graph.appendChild(canvas);

        const status = document.createElement("div");
        status.className = "list-status";
        div.appendChild(status);

        const remove = document.createElement("div");
        remove.className = "list-remove";
        div.appendChild(remove);
        
        remove.onclick = () => { this.Remove(host); };

        let history = [];
        for (let i = 0; i < PING_HISTORY_LEN; i++) {
            history.push(-1);
        }

        this.hashtable[host] = {
            host: host,
            element: div,
            status: status,
            graph: graph,
            canvas: canvas,
            history: history
        };

        this.args.entries.push(host);
        
        this.txtInput.style.left = "8px";
        this.txtInput.style.bottom = "8px";
        this.txtInput.style.width = "calc(100% - 16px)";

        if (this.ws != null && this.ws.readyState === 0) { //connection

        } else if (this.ws != null && this.ws.readyState === 1) { //ready
            this.ws.send("add:" + host);

        } else {
            this.Connect();
        }
    }

    Remove(host) {
        if (this.hashtable[host]) {
            this.list.removeChild(this.hashtable[host].element);
            delete this.hashtable[host];
        }

        let index = this.args.entries.indexOf(host);
        if (index > -1) {
            this.args.entries.splice(index, 1);
        }
        
        if (this.ws.readyState === 1) {
            this.ws.send("remove:" + host);
            if (this.args.entries.length == 0) this.ws.close();
        }

        this.AfterResize();
    }

    Connect() {
        let server = window.location.href;
        server = server.replace("https://", "");
        server = server.replace("http://", "");
        if (server.indexOf("/") > 0) server = server.substring(0, server.indexOf("/"));

        if (this.ws != null)
            try {
                this.ws.close();
            } catch (error) { };

        this.ws = new WebSocket((isSecure ? "wss://" : "ws://") + server + "/ws/ping");

        this.ws.onopen = () => {
            this.ws.send("timeout:" + this.args.timeout);
            this.ws.send("method:" + this.args.method);

            let i = 0;
            while (i < this.args.entries.length) {
                let req = "add:";
                while (req.length < 768 && i < this.args.entries.length) {
                    if (this.args.entries[i].length > 0) req += this.args.entries[i] + ";";
                    i++;
                }
                this.ws.send(req);
            }
        };

        this.ws.onclose = () => {

        };

        this.ws.onmessage = event => {

        };

        //this.ws.onerror = error => { console.log(error); };
    }

    DrawGraph(host) {

    }

    InvalidateRecyclerList() { //override
        for (let key in this.hashtable)
            if (this.hashtable[key].element.offsetTop - this.list.scrollTop < -30 ||
                this.hashtable[key].element.offsetTop - this.list.scrollTop > this.list.clientHeight) {
                this.hashtable[key].graph.style.display = "none";
            } else {
                this.hashtable[key].graph.style.display = "initial";
            }
    }

}
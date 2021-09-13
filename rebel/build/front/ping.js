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

        this.SetTitle(this.args.method == "arp" ? "ARP ping" : "Ping");
        this.SetIcon("res/ping.svg");

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
        const div = document.createElement("div");
        div.className = "list-element";
        this.list.appendChild(div);

        const name = document.createElement("div");
        name.className = "list-label";
        name.innerHTML = hostname;
        div.appendChild(name);

        const graph = document.createElement("div");
        graph.className = "list-graph";
        div.appendChild(graph);

        const msg = document.createElement("div");
        msg.className = "list-msg";
        div.appendChild(msg);

        const remove = document.createElement("div");
        remove.className = "list-remove";
        div.appendChild(remove);
    }

    Remove(host) {

    }


}
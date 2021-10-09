package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	PROTOCOL_ICMP   = 1
	PROTOCOL_ICMPv6 = 58
)

func IcmpRequestAndResolve(addr string, timeout int, ttl int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := net.ResolveIPAddr("ip4", addr)
	if err == nil {
		return IcmpRequest(ipaddr, timeout, ttl, seq)
	}
	return nil, 0, err
}

func IcmpRequest(ipaddr *net.IPAddr, timeout int, ttl int, seq int) (*net.IPAddr, float32, error) {
	//listening for icmp replies
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, 0, err
	}
	conn.IPv4PacketConn().SetTTL(ttl)
	defer conn.Close()

	//icmp message
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte(""),
		},
	}
	b, err := message.Marshal(nil)
	if err != nil {
		return ipaddr, 0, err
	}

	//send it
	var start time.Time = time.Now()
	n, err := conn.WriteTo(b, ipaddr)
	if err != nil {
		return ipaddr, 0, err
	} else if n != len(b) {
		return ipaddr, 0, errors.New("buffer mismatch")
	}

	var deadline time.Duration = time.Duration(timeout) * time.Millisecond

	//wait for a reply
	var reply []byte = make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(deadline))
	if err != nil {
		return ipaddr, 0, err
	}

	n, peer, err := conn.ReadFrom(reply)
	var rtt time.Duration = time.Since(start)
	if err != nil {
		if rtt > deadline {
			return ipaddr, float32(deadline / time.Millisecond), errors.New("timed out")
		}
		return ipaddr, 0, err
	}

	//done
	rm, err := icmp.ParseMessage(PROTOCOL_ICMP, reply[:n])
	if err != nil {
		return ipaddr, 0, err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return ipaddr, float32(rtt.Nanoseconds()) / 1000000.0, nil

	default:
		return ipaddr, 0, fmt.Errorf("%+v from %v", rm, peer)
	}
}

func IcmpRequestV6AndResolve(addr string, timeout int, ttl int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := net.ResolveIPAddr("ip6", addr)
	if err == nil {
		return IcmpRequestV6(ipaddr, timeout, ttl, seq)
	}
	return nil, 0, err
}

func IcmpRequestV6(ipaddr *net.IPAddr, timeout int, ttl int, seq int) (*net.IPAddr, float32, error) {
	//listening for icmp replies
	conn, err := icmp.ListenPacket("ip6:ipv6-icmp", "::0")
	if err != nil {
		return nil, 0, err
	}
	conn.IPv6PacketConn().SetHopLimit(64)
	defer conn.Close()

	//icmp message
	message := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte(""),
		},
	}
	b, err := message.Marshal(nil)
	if err != nil {
		return ipaddr, 0, err
	}

	//send it
	var start time.Time = time.Now()
	n, err := conn.WriteTo(b, ipaddr)
	if err != nil {
		return ipaddr, 0, err
	} else if n != len(b) {
		return ipaddr, 0, errors.New("buffer mismatch")
	}

	var deadline time.Duration = time.Duration(timeout) * time.Millisecond

	//wait for a reply
	var reply []byte = make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(deadline))
	if err != nil {
		return ipaddr, 0, err
	}

	n, peer, err := conn.ReadFrom(reply)
	var rtt time.Duration = time.Since(start)
	if err != nil {
		if rtt > deadline {
			return ipaddr, float32(deadline / time.Millisecond), errors.New("timed out")
		}
		return ipaddr, 0, err
	}

	//done
	rm, err := icmp.ParseMessage(PROTOCOL_ICMPv6, reply[:n])
	if err != nil {
		return ipaddr, 0, err
	}

	switch rm.Type {
	case ipv6.ICMPTypeEchoReply:
		return ipaddr, float32(rtt.Nanoseconds()) / 1000000.0, nil

	default:
		return ipaddr, 0, fmt.Errorf("%+v from %v", rm, peer)
	}
}

func WsPing(w http.ResponseWriter, r *http.Request) {
	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		return CheckOrigin(r)
	}

	var n func() = func() {
		println("")
	}
	n()

	ws, wsErr := wsUpgrader.Upgrade(w, r, nil)
	if wsErr != nil {
		log.Println(wsErr)
		return
	}

	var closed bool = false

	var hosts []string
	var seq int = 0
	var ttl int = 64
	var timeout int = 1000
	var interval int = 1000
	var method string = "icmp"
	//var forceIpv6 bool = false

	go func() { //ping loop
		for !closed {
			var results []string = make([]string, len(hosts))

			var startTime time.Time = time.Now()

			var wg sync.WaitGroup
			for i := 0; i < len(hosts); i++ {
				wg.Add(1)
				go func(i int, host string) {
					defer wg.Done()

					var ipaddr *net.IPAddr
					var rtt float32
					var icmpErr error
					ipaddr, rtt, icmpErr = IcmpRequestAndResolve(host, timeout, ttl, seq)

					if icmpErr == nil {
						results[i] = "\"" + host + "\":[\"" + ipaddr.String() + "\",\"" + fmt.Sprint(rtt) + "\"]"
					} else {
						results[i] = "\"" + host + "\":[\"" + icmpErr.Error() + "\"]"
					}

				}(i, hosts[i])
			}
			wg.Wait()

			if writeErr := ws.WriteMessage(WEBSOCKET_TYPE_BINARY, []byte("{\"o\":{"+strings.Join(results, ",")+"}}")); writeErr != nil {
				closed = true
				log.Println(writeErr)
				return
			}

			var endTime time.Time = time.Now()

			var sleepTime = int64(interval) - endTime.Sub(startTime).Milliseconds()
			if sleepTime > 0 {
				time.Sleep(time.Duration(sleepTime * 1000000))
			}

			seq++
		}
	}()

	defer func() {
		closed = true
	}()

	for !closed { //communication loop
		_, bytes, readErr := ws.ReadMessage()
		if readErr != nil {
			closed = true
			return
		}

		var msg []string = strings.Split(string(bytes), ":")

		if len(msg) < 2 {
			continue
		}
		msg[0] = strings.TrimSpace(msg[0])

		println(msg[0], msg[1])

		switch msg[0] {
		case "add":
			seq = 0
			var split []string = strings.Split(msg[1], ";")

			for _, e := range split {
				if len(e) == 0 {
					continue
				}

				var exist bool = false
				for _, e2 := range hosts {
					if e2 == e {
						exist = true
						break
					}
				}

				if !exist {
					hosts = append(hosts, e)
				}
			}

		case "remove":
			var index int = -1
			for i := 0; i < len(hosts); i++ {
				if hosts[i] == msg[1] {
					index = i
					break
				}
			}

			if index > -1 {
				hosts = append(hosts[:index], hosts[index+1:]...)
			}

		case "timeout":
			if num, err := strconv.Atoi(msg[1]); err == nil {
				timeout = num
			}

		case "interval":
			if num, err := strconv.Atoi(msg[1]); err == nil {
				interval = num
			}

		case "ttl":
			if num, err := strconv.Atoi(msg[1]); err == nil {
				ttl = num
			}

		case "method":
			method = msg[1]

			//case "forceipv6":
			//	forceIpv6 = (msg[1] == "true")
		}

		println(ttl, timeout)

		if false {
			println(msg, method)
		}

		if closed {
			return
		}

	}

}

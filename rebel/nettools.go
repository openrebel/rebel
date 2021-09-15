package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	PROTOCOL_ICMP   = 1
	PROTOCOL_ICMPv6 = 58
)

func DnsLookup(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func IcmpRequestAndResolve(addr string, timeout int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := net.ResolveIPAddr("ip4", addr)
	if err == nil {
		return IcmpRequest(ipaddr, timeout, seq)
	}
	return nil, 0, err
}

func IcmpRequest(ipaddr *net.IPAddr, timeout int, seq int) (*net.IPAddr, float32, error) {
	//listening for icmp replies
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, 0, err
	}
	conn.IPv4PacketConn().SetTTL(64)
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

func IcmpRequestV6AndResolve(addr string, timeout int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := net.ResolveIPAddr("ip6", addr)
	if err == nil {
		return IcmpRequestV6(ipaddr, timeout, seq)
	}
	return nil, 0, err
}

func IcmpRequestV6(ipaddr *net.IPAddr, timeout int, seq int) (*net.IPAddr, float32, error) {
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

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var closed bool = false

	var hosts []string
	//var forceIpv6 bool = false
	//var ttl byte = 64
	var timeout int = 1000
	var seq int = 0
	var method string = "icmp"

	go func() { //ping loop
		for !closed {

			if method == "icmp" {
				for i := 0; i < len(hosts); i++ {
					IcmpRequestAndResolve(hosts[i], timeout, seq)
				}
			}

			time.Sleep(time.Duration(timeout))
		}
	}()

	for { //communication loop
		messageType, bytes, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		//log.Println(string(bytes))

		if err := ws.WriteMessage(messageType, bytes); err != nil {
			log.Println(err)
			return
		}
	}

}

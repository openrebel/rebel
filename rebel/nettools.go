package main

import (
	"errors"
	"fmt"
	"net"
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

func DnsResolve(addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr("ip4", addr)
}

func DnsResolveIPv6(addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr("ip6", addr)
}

func DnsLookup(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func IcmpRequest(addr string, timeout int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := DnsResolve(addr)
	if err != nil {
		return nil, 0, err
	}

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
		return ipaddr, 0, fmt.Errorf("got %v; want %v", n, len(b))
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
		return ipaddr, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

func IcmpRequestV6(addr string, timeout int, seq int) (*net.IPAddr, float32, error) {
	ipaddr, err := DnsResolveIPv6(addr)
	if err != nil {
		return nil, 0, err
	}

	//listening for icmp replies
	println("list")
	conn, err := icmp.ListenPacket("ip6:ipv6-icmp", "::0")
	if err != nil {
		return nil, 0, err
	}
	conn.IPv6PacketConn().SetHopLimit(64)
	defer conn.Close()

	//icmp message
	println("msg")
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
	println("sent it")
	var start time.Time = time.Now()
	n, err := conn.WriteTo(b, ipaddr)
	if err != nil {
		return ipaddr, 0, err
	} else if n != len(b) {
		return ipaddr, 0, fmt.Errorf("got %v; want %v", n, len(b))
	}

	var deadline time.Duration = time.Duration(timeout) * time.Millisecond

	//wait for a reply
	println("reply")
	var reply []byte = make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(deadline))
	if err != nil {
		return ipaddr, 0, err
	}

	println("read")
	n, peer, err := conn.ReadFrom(reply)
	var rtt time.Duration = time.Since(start)
	if err != nil {
		if rtt > deadline {
			return ipaddr, float32(deadline / time.Millisecond), errors.New("timed out")
		}
		return ipaddr, 0, err
	}

	//done
	println("done?")
	rm, err := icmp.ParseMessage(PROTOCOL_ICMPv6, reply[:n])
	if err != nil {
		return ipaddr, 0, err
	}

	switch rm.Type {
	case ipv6.ICMPTypeEchoReply:
		return ipaddr, float32(rtt.Nanoseconds()) / 1000000.0, nil

	default:
		return ipaddr, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

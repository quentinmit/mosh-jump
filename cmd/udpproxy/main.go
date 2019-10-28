package main

import (
	"flag"
	"log"
	"net"
)

var (
	laddr = flag.String("laddr", "", "listen address")
	raddr = flag.String("raddr", "", "remote server address")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	udpLAddr, err := net.ResolveUDPAddr("udp", *laddr)
	if err != nil {
		return err
	}
	udpRAddr, err := net.ResolveUDPAddr("udp", *raddr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpLAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	remoteConnections := make(map[string]*net.UDPConn)
	for {
		pkt := make([]byte, 1500)
		n, ua, err := conn.ReadFromUDP(pkt)
		if err != nil {
			return err
		}
		pkt = pkt[:n]
		rconn := remoteConnections[ua.String()]
		if rconn == nil {
			rconn, err = net.DialUDP("udp", nil, udpRAddr)
			if err != nil {
				return err
			}
			remoteConnections[ua.String()] = rconn
			go readLoop(conn, rconn, ua)
		}
		if _, err := rconn.Write(pkt); err != nil {
			return err
		}
	}
}

// readLoop reads packets from rconn (the server) and sends them to ua on lconn (the client).
func readLoop(lconn, rconn *net.UDPConn, ua *net.UDPAddr) error {
	for {
		pkt := make([]byte, 1500)
		n, err := rconn.Read(pkt)
		if err != nil {
			return err
		}
		pkt = pkt[:n]
		if _, err := lconn.WriteTo(pkt, ua); err != nil {
			return err
		}
	}
}

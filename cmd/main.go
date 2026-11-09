package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

	"M-Faheem-Khan/octo-fiesta/pkg/iface"
	"M-Faheem-Khan/octo-fiesta/pkg/packet"
	"M-Faheem-Khan/octo-fiesta/pkg/protocol"
)

// getInterface: Create, Configures & return fd for interface
func getInterface(ifaceName, cidr string) (*os.File, string) {
	fd, ifname, err := iface.OpenTun(ifaceName, false)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created Interface: %s", ifname)
	log.Printf("Configuring Interface: %s", ifname)

	ipAddr, err := iface.SetUpIface(ifaceName, cidr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Interface(%s) assigned IP: '%s'", ifname, ipAddr)
	return fd, ipAddr
}

func main() {
	// read Mode flag(server/client)
	// Once the Server is started it will share its IP
	// Once the Client is started it will init - ICMP Echo check.

	var mode, ifaceName, serverIp string
	flag.StringVar(&mode, "mode", "server", "Running Mode(default: server).")
	flag.StringVar(&ifaceName, "interface", "tun0", "Interface Name - (defualt: tun0).")
	flag.StringVar(&serverIp, "serverIp", "10.0.0.2", "Server IP - (defualt: 10.0.0.2).")
	flag.Parse()

	switch mode {
	case "server":
		log.Printf("Running in %s mode.\n", mode)
		tunFd, _ := getInterface(ifaceName, "10.0.0.1/24")
		buf := make([]byte, 2048)
		for {
			stream, err := tunFd.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			// ICMP Reply
			// Verify if the packet is valid
			if stream >= 20 && (buf[0]>>4) == 4 && buf[9] == 1 {
				packet := buf[:stream]
				// verify: ICMP Echo Request(Type 8)
				ipHeaderLen := int(buf[0]&0x0F) * 4
				if len(packet) >= ipHeaderLen+8 && packet[ipHeaderLen] == 8 {
					icmpReply := protocol.BuildIcmpReply(packet, ipHeaderLen)
					_, err := tunFd.Write(icmpReply)
					if err != nil {
						log.Printf("Error writing reply to tun0: %v", err)
					} else {
						srcIP := net.IPv4(packet[12], packet[13], packet[14], packet[15])
						dstIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])
						log.Printf("[PING RESPONDED] Replied to ping from %s (Originally sent to %s) | %d bytes, payload: %q\n",
							srcIP, dstIP, len(icmpReply), packet[ipHeaderLen+8:])
					}
				}
			}
			packet.InspectPacket(buf[:stream])
		}
	case "client":
		// The interface address (10.0.1.1) must differ from the packet source
		// (10.0.1.2): packets from a local address are dropped as martians,
		// and the kernel only forwards packets from a non-local source.
		tunFd, _ := getInterface(ifaceName, "10.0.1.1/24")
		const clientSrcIp = "10.0.1.2"
		icmpPkt := protocol.BuildIcmpRequest(clientSrcIp, serverIp)

		// Read echo replies forwarded back to the client interface
		go func() {
			rbuf := make([]byte, 2048)
			for {
				n, err := tunFd.Read(rbuf)
				if err != nil {
					log.Fatal(err)
				}
				if n >= 20 && (rbuf[0]>>4) == 4 && rbuf[9] == 1 {
					ihl := int(rbuf[0]&0x0F) * 4
					if n >= ihl+8 && rbuf[ihl] == 0 {
						src := net.IPv4(rbuf[12], rbuf[13], rbuf[14], rbuf[15])
						log.Printf("[ICMP Echo Reply] Received from %s | %d bytes, payload: %q\n",
							src, n, rbuf[ihl+8:n])
					}
				}
			}
		}()

		for {
			n, err := tunFd.Write(icmpPkt)
			if err != nil {
				log.Printf("Error writing request to %s: %v", ifaceName, err)
			} else {
				log.Printf("[ICMP Echo Request] Sent to %s | %d bytes, payload: %q\n",
					serverIp, n, icmpPkt[28:])
			}
			time.Sleep(time.Second) // sleep for 1s
		}
	default:
		log.Printf("Error: Invalid mode(server/client), '%s' not defined.\n", mode)
		os.Exit(1)
	}
}

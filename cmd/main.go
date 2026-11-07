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

const ifaceName = "tun0"

// getInterface: Create, Configures & return fd for interface
func getInterface() (*os.File, string) {
	fd, ifname, err := iface.OpenTun(ifaceName, false)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created Interface: %s", ifname)
	log.Printf("Configuring Interface: %s", ifname)

	ipAddr, err := iface.SetUpIface(ifaceName, "10.0.0.1/24")
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
		tunFd, _ := getInterface()
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
						log.Printf("[PING RESPONDED] Replied to ping from %s (Originally sent to %s)\n", srcIP, dstIP)
					}
				}
			}
			packet.InspectPacket(buf[:stream])
		}
	case "client":
		log.Printf("Running in %s mode.\n", mode)
		tunFd, srcIp := getInterface()
		for {
			icmpPkt := protocol.BuildIcmpRequest(srcIp, serverIp)
			_, err := tunFd.Write(icmpPkt)
			if err != nil {
				log.Printf("Error writing reply to tun0: %v", err)
			} else {
				log.Printf("[ICMP Echo Request] Sent to %s\n")
			}
			time.Sleep(time.Second) // sleep for 1s
		}
	default:
		log.Printf("Error: Invalid mode(server/client), '%s' not defined.\n", mode)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	f, name, err := OpenTun("tun9", false)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	log.Printf("Created %s", name)

	if err := SetUpIface(name, "10.0.0.1/24"); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s is up with 10.0.0.1/24", name)

	buf := make([]byte, 2048)
	for {
		n, err := f.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read %d bytes: % x.\n", n, buf[:min(n, 32)])
		// InspectPacket(buf[:n])

		// ICMP Reply
		// Verify if the packet is valid
		if n >= 20 && (buf[0]>>4) == 4 && buf[9] == 1 {
			packet := buf[:n]
			// verify: ICMP Echo Request(Type 8)
			ipHeaderLen := int(buf[0]&0x0F) * 4
			if len(packet) >= ipHeaderLen+8 && packet[ipHeaderLen] == 8 {
				icmpReply := BuildIcmpReply(packet, ipHeaderLen)
				_, err := f.Write(icmpReply)
				if err != nil {
					log.Printf("Error writing reply to tun0: %v", err)
				} else {
					srcIP := net.IPv4(packet[12], packet[13], packet[14], packet[15])
					dstIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])
					fmt.Printf("[PING RESPONDED] Replied to ping from %s (Originally sent to %s)\n", srcIP, dstIP)
				}
			}
		}
	}
}

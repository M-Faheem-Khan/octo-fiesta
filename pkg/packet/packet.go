package packet

import "fmt"

func InspectPacket(packet []byte) {
	// Verify the packet meets the min size
	if len(packet) <= 0 {
		fmt.Printf("Not a valid IPv4 Packet(20 - 65355): %d.\n", len(packet))
		return
	}

	// Parse IP Header
	// Verions, Internet Header Length(IHL)
	versionAndIHL := packet[0]
	version := versionAndIHL >> 4
	if version != 4 {
		fmt.Printf("Not a IPv4 packet: IPv%d", version)
		return
	}

	ihl := int(versionAndIHL & 0x0F) // IHL in 32-bit WORDS
	ipHeaderLen := ihl * 4           // total length in bytes

	totalLength := int(packet[2])<<8 | int(packet[3])
	protocol := packet[9] // protocol number(1=ICMP)
	srcIp := [4]byte{packet[12], packet[13], packet[14], packet[15]}
	dstIp := [4]byte{packet[16], packet[17], packet[18], packet[19]}

	if len(packet) < ipHeaderLen {
		return
	}

	transportPayload := packet[ipHeaderLen:]

	if protocol == 1 {
		if len(transportPayload) >= 8 {
			icmpType := transportPayload[0]
			icmpCode := transportPayload[1]
			icmpData := transportPayload[8:]

			fmt.Printf("[ICMP Packet] %s -> %s | Type: %d, Code: %d | Payload Size: %d\n",
				srcIp, dstIp, icmpType, icmpCode, len(icmpData),
			)
		}
	} else if protocol == 6 { // TCP
		fmt.Println("TCP Packet")
	} else if protocol == 17 { // UDP
		fmt.Println("UDP Packet")
	} else {
		fmt.Printf("[Unknown Protocol %d] %s -> %s | Total Len: %d\n", protocol, srcIp, dstIp, totalLength)
	}

}

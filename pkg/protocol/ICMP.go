package protocol

import "encoding/binary"

func BuildIcmpRequest(srcIp, dstIp string) []byte {
	// 0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Version|  IHL  |Type of Service|          Total Length         |
	// | (0x4) | (0x5) |    (0x00)     |          (0x003B = 59)        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |         Identification        |Flags|     Fragment Offset     |
	// |         (0x3039 = 12345)      |0 1 0|       (0x0000)        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  Time to Live |   Protocol    |       Header Checksum         |
	// |    (0x40)     |  (0x01=ICMP)  |           (0x0000)            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                       Source IP Address                       |
	// |                          (10.0.0.1)                           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                    Destination IP Address                     |
	// |                          (10.0.0.2)                           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |   ICMP Type   |   ICMP Code   |        ICMP Checksum          |
	// |  (0x08=Req)   |    (0x00)     |       [Checksum Bytes]        |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |          Identifier           |        Sequence Number        |
	// |       (0x1111 = 4369)         |        (0x0001 = Seq 1)       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                 Payload: "Hello ICMP"                         |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	// IP Header -> 20 bytes
	// ICMP Header -> 8 bytes
	// Payload("Hello ICMP") -> 10 bytes
	// IHL = IP header + ICMP header + payload = 38

	payload := []byte("Hello ICMP")
	ipHeaderLength := 20
	icmpHeaderLength := 8
	totalLength := len(payload) + ipHeaderLength + icmpHeaderLength
	packet := make([]byte, totalLength) // allocating packet
	packet[0] = 0x45                    // version, IHL
	packet[1] = 0x00                    // type of service(TOS)
	// total length
	binary.BigEndian.PutUint16(packet[2:4], uint16(totalLength))
	// identification: set by sender in our case it doesn't
	// matter as we are only sending a single packet)
	binary.BigEndian.PutUint16(packet[5:6], uint16(12345))
	// Flags: Don't fragment flag set(010)
	// flags: reserved|don't fragment|more fragment
	// flag(0x4) + fragment(0x0000) -> 0x4000 -> 0100 0000 0000 0000
	// flag(3 bits) + fragment(13 bits)
	binary.BigEndian.PutUint16(packet[6:8], 0x4000)
	packet[8] = 64 // TTL -> 64 -> 01000000 -> 0x40
	packet[9] = 1  // Protocol(ICMP)

	// setting checksum to 0
	binary.BigEndian.PutUint16(packet[10:12], 0x0000)

	// Source IP (bytes 12-15)
	copy(packet[12:16], srcIp)
	// Destination IP (bytes 16-19)
	copy(packet[16:20], dstIp)

	// Building ICMP Header
	icmpHeader := packet[20:]
	icmpHeader[0] = 8 // Type = 8 (Echo Request)
	icmpHeader[1] = 0 // Code = 0

	// Checksum bytes (2-3 of ICMP header) temporarily set to 0 before calculation
	icmpHeader[2] = 0
	icmpHeader[3] = 0

	// Identifier: in our case doesn't matter (for tracking)
	binary.BigEndian.PutUint16(icmpHeader[4:6], 0xFFFF)
	// Sequence: in our case doesn't matter
	binary.BigEndian.PutUint16(icmpHeader[6:8], 1)

	// Copy payload data right after the 8-byte ICMP header
	copy(icmpHeader[8:], payload)

	csum := calculateChecksum(icmpHeader)
	binary.BigEndian.PutUint16(icmpHeader[2:4], csum)

	return packet
}

func BuildIcmpReply(packet []byte, ipHeaderLength int) []byte {
	reply := make([]byte, len(packet))
	copy(reply, packet)

	icmpHeader := reply[ipHeaderLength:]

	// Change ICMP Reply Code(8 -> Echo Request, 0 -> Echo Response)
	icmpHeader[0] = 0

	// Swap srcIp(12 - 16) w/ dstIp(16 - 20)
	srcIp := [4]byte{packet[12], packet[13], packet[14], packet[15]}
	dstIp := [4]byte{packet[16], packet[17], packet[18], packet[19]}
	copy(reply[12:16], dstIp[:])
	copy(reply[16:20], srcIp[:])

	// Clear Checksum
	icmpHeader[2] = 0
	icmpHeader[3] = 0

	// Calculate Checksum(required)
	checksum := calculateChecksum(icmpHeader)
	icmpHeader[2] = byte(checksum >> 8)
	icmpHeader[3] = byte(checksum & 0xFF)

	return reply
}

func calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

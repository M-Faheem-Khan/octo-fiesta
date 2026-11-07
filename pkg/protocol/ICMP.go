package protocol

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

package main

import (
	"log"
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
		InspectPacket(buf[:n])
	}
}

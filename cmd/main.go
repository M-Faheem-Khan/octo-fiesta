package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"M-Faheem-Khan/octo-fiesta/pkg/iface"
)

const ifaceName = "tun0"

// getInterface: Create, Configures & return fd for interface
func getInterface() *os.File {
	fd, ifname, err := iface.OpenTun(ifaceName, false)
	if err != nil {
		log.Fatal(err)
	}

	defer fd.Close()

	log.Printf("Created Interface: %s", ifname)
	log.Printf("Configuring Interface: %s", ifname)

	ipAddr, err := iface.SetUpIface(ifaceName, "10.0.0.1/24")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Interface(%s) assigned IP: %s", ifname, ipAddr)
	return fd
}

func main() {
	// read Mode flag(server/client)
	// Once the Server is started it will share its IP
	// Once the Client is started it will init - ICMP Echo check.

	var mode, ifaceName string
	flag.StringVar(&mode, "mode", "server", "Running Mode(default: server).")
	flag.StringVar(&ifaceName, "interface", "tun0", "Interface Name - (defualt: tun0).")
	flag.Parse()

	switch mode {
	case "server":
		fmt.Printf("Running in %s mode.\n", mode)
		getInterface()
	case "client":
		fmt.Printf("Running in %s mode.\n", mode)
	default:
		fmt.Printf("Error: Invalid mode(server/client), '%s' not defined.\n", mode)
		os.Exit(1)
	}
}

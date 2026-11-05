## Creating a Tunnel/Tap interface & Read(RAW) data.

The section covers how to create, configure & read incoming packets from an interface. For simplicity the target system is linux(alpine) running inside of container(docker). 


## Creating a new Interface
Before we can create an interface we must define what an interface is. In this instance an interface is a point of connection between a system(OS) and network that allows the interface(Wifi card, Ethernet card) to send/recieve data over the network(LAN, WAN, etc.).

> [!NOTE]  
> Ensure that `/dev/net/tun` is accessible(root, NET_ADMIN capability, access to device(`/dev/net/tun`)). This is already configured in the docker setup.


We only need `Name` & `Flags` for creating the interface. The `Name` is the name of the interface and `Flags` for creating the interface. We will use three flags for creating the intial interface. In the next steps we will use additional flags to configure the interface.
- IFF_TAP: Layer 2 - Data Link Layer: Defines the data format on the network() - Access Point
- IFF_TUN: Layer 3 - Network Layer: RAW Network Packets(IPv4/IPv6) - Full Tunnel
- IFF_NO_PI: Do not send Packet Information(Mac addr etc.)

> [!NOTE]
> Kernel expects the whole 40 bytes. See [man page: linux/netdevice](https://man7.org/linux/man-pages/man7/netdevice.7.html) for entire struct format. 

```Golang
type ifReqFlags struct {
    Name  [unix.IFNAMSIZ]byte
    Flags uint16
    _     [22]byte // pad to 40 bytes
}
```

```Golang
func openTun(ifName string, tap bool) (*os.File, string, error) {
    // Open /dev/net/tun
    tunFd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
    // ...

	var req ifReqFlags
	copy(req.Name[:], []byte(ifName)) // name of the new interface
	if tap {
		req.Flags = unix.IFF_TAP | unix.IFF_NO_PI // TAP without Packet Information
	} else {
		req.Flags = unix.IFF_TUN | unix.IFF_NO_PI // Tunnel without Packet Information
	}

    // TUNSETIFF - Tunnel Set Interface Flags
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(tunFd),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&req)), // flags to apply
	)
    // ...

	return os.NewFile(uintptr(tunFd), "/dev/net/tun"), string(req.Name[:n]), nil
}
```

## Assigning IP address & Network Mask
```Golang
type ifReqAddr struct {
	Name   [unix.IFNAMSIZ]byte
	Family uint16
	Port   uint16
	Addr   [4]byte
	_      [16]byte // pad to 40 bytes
}

// IOCTL Syscall Helper func
func ioctl(fd int, op uintptr, arg unsafe.Pointer) error {
	if _, _, e := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), op, uintptr(arg)); e != 0 {
		return e
	}
	return nil
}

func setupIface(name, cidr, string) error {
    // Get IPv4 & Netmask
    ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("parse cidr: %w", err)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return fmt.Errorf("%q is not an IPv4 address", cidr)
	}
	mask := ipnet.Mask // 4 bytes for an IPv4 CIDR

    s, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0) // open socket
    //...

    // 1. Set IP
    // 2. Set Network mask
    // 3. Bring up the Link
        // get existing flags
        // update flags

    	// 1. Set the address.
	var ar ifReqAddr
	copy(ar.Name[:], name)
	ar.Family = unix.AF_INET
	copy(ar.Addr[:], ip4)
	if err := ioctl(s, unix.SIOCSIFADDR, unsafe.Pointer(&ar)); err != nil {
		return fmt.Errorf("SIOCSIFADDR: %w", err)
	}

	// 2. Set the netmask.
	var mr ifReqAddr
	copy(mr.Name[:], name)
	mr.Family = unix.AF_INET
	copy(mr.Addr[:], mask)
	if err := ioctl(s, unix.SIOCSIFNETMASK, unsafe.Pointer(&mr)); err != nil {
		return fmt.Errorf("SIOCSIFNETMASK: %w", err)
	}

	// 3. Bring the link up (read-modify-write the flags).
	var fr ifReqFlags
	copy(fr.Name[:], name)
	if err := ioctl(s, unix.SIOCGIFFLAGS, unsafe.Pointer(&fr)); err != nil {
		return fmt.Errorf("SIOCGIFFLAGS: %w", err)
	}
	
    fr.Flags |= unix.IFF_UP | unix.IFF_RUNNING // update the flags
	
    if err := ioctl(s, unix.SIOCSIFFLAGS, unsafe.Pointer(&fr)); err != nil {
		return fmt.Errorf("SIOCSIFFLAGS: %w", err)
	}
}
```

## Closing the Loop
We will now close the loop by creating, configuring and listing on the interface.

```Golang
func main() {
	f, name, err := openTun("tun9", false)
    //...
	log.Printf("Created %s", name)

	if err := setUpIface(name, "10.0.0.1/24"); err != nil {
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
	}
}
```

## Setting up Docker
```YAML
# Build Stage
FROM golang:bookworm AS builder
RUN mkdir /build
COPY . /build/
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o tunnel cmd/main.go

# Deploy Stage
FROM alpine:3.20
RUN apk add --no-cache iproute2
COPY --from=builder /build/tunnel /usr/local/bin/tunnel
ENTRYPOINT [ "/usr/local/bin/tunnel" ]
```

```Yaml
services:
  tunnel:
    build: .
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun:/dev/net/tun
```

## Testing
- Run `docker compose up --build`  
- Exec into the container in a new shell: `docker exec -it tunnel sh`
- Send Ping: `ping 10.0.0.2`
- See the incoming data in the other first terminal.



<!-- EOF -->

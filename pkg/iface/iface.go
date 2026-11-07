package iface

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func OpenTun(ifName string, tap bool) (*os.File, string, error) {
	// Open Tunnel FD
	tunFd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, "", fmt.Errorf("open /dev/net/tun: %w", err)
	}

	var req ifReqFlags
	copy(req.Name[:], []byte(ifName))
	if tap {
		req.Flags = unix.IFF_TAP | unix.IFF_NO_PI
	} else {
		req.Flags = unix.IFF_TUN | unix.IFF_NO_PI
	}

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(tunFd),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&req)),
	)

	if errno != 0 {
		unix.Close(tunFd)
		return nil, "", fmt.Errorf("TUNSETIFF: %w", errno)
	}

	// Kernel may have filled in the real name
	n := 0
	for n < len(req.Name) && req.Name[n] != 0 {
		n++
	}

	return os.NewFile(uintptr(tunFd), "/dev/net/tun"), string(req.Name[:n]), nil
}

func ioctl(fd int, op uintptr, arg unsafe.Pointer) error {
	if _, _, e := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), op, uintptr(arg)); e != 0 {
		return e
	}
	return nil
}

// setUp assigns an IPv4 address/netmask (e.g. "10.0.0.1/24") to the
// interface and brings it up.
func SetUpIface(name, cidr string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("parse cidr: %w", err)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return "", fmt.Errorf("%q is not an IPv4 address", cidr)
	}
	mask := ipnet.Mask // 4 bytes for an IPv4 CIDR

	s, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return "", fmt.Errorf("socket: %w", err)
	}
	defer unix.Close(s)

	// 1. Set the address.
	var ar ifReqAddr
	copy(ar.Name[:], name)
	ar.Family = unix.AF_INET
	copy(ar.Addr[:], ip4)
	if err := ioctl(s, unix.SIOCSIFADDR, unsafe.Pointer(&ar)); err != nil {
		return "", fmt.Errorf("SIOCSIFADDR: %w", err)
	}

	// 2. Set the netmask.
	var mr ifReqAddr
	copy(mr.Name[:], name)
	mr.Family = unix.AF_INET
	copy(mr.Addr[:], mask)
	if err := ioctl(s, unix.SIOCSIFNETMASK, unsafe.Pointer(&mr)); err != nil {
		return "", fmt.Errorf("SIOCSIFNETMASK: %w", err)
	}

	// 3. Bring the link up (read-modify-write the flags).
	var fr ifReqFlags
	copy(fr.Name[:], name)
	if err := ioctl(s, unix.SIOCGIFFLAGS, unsafe.Pointer(&fr)); err != nil {
		return "", fmt.Errorf("SIOCGIFFLAGS: %w", err)
	}
	fr.Flags |= unix.IFF_UP | unix.IFF_RUNNING
	if err := ioctl(s, unix.SIOCSIFFLAGS, unsafe.Pointer(&fr)); err != nil {
		return "", fmt.Errorf("SIOCSIFFLAGS: %w", err)
	}

	// Get the IP
	if err := ioctl(s, unix.SIOCGIFADDR, unsafe.Pointer(&ar)); err != nil {
		return "", fmt.Errorf("SIOCSIFADDR: %w", err)
	}

	// Return IP
	return netip.AddrFrom4(ar.Addr).String(), nil
}

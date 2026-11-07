package iface

import "golang.org/x/sys/unix"

type ifReqFlags struct {
	Name  [unix.IFNAMSIZ]byte
	Flags uint16
	_     [22]byte // pad to 40 bytes
}

type ifReqAddr struct {
	Name   [unix.IFNAMSIZ]byte
	Family uint16
	Port   uint16
	Addr   [4]byte
	_      [16]byte // pad to 40 bytes
}

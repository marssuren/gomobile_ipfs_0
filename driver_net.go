package main

import "net"

type NativeNetDriver interface {
	InterfaceAddrs() (*NetAddrs, error)
	Interfaces() (*NetInterfaces, error)
}

type NetAddrs struct {
	addrs []string
}

type NetInterfaces struct {
	ifaces []*NetInterface
}

type NetInterface struct {
	Index int       // positive integer that starts at one, zero is never used
	MTU   int       // maximum transmission unit
	Name  string    // e.g., "en0", "lo0", "eth0.100"
	Addrs *NetAddrs // InterfaceAddresses

	hardwareaddr []byte    // IEEE MAC-48, EUI-48 and EUI-64 form
	flags        net.Flags // e.g., FlagUp, FlagLoopback, FlagMulticast
}

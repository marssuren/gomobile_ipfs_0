package main

type NativeMDNSLockerDriver interface {
	Lock()
	Unlock()
}

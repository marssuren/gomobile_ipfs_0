package core

type NativeMDNSLockerDriver interface {
	Lock()
	Unlock()
}

package core

type NodeConfig struct {
	bleDriver        ProximityDriver
	netDriver        NativeNetDriver
	mdnsLockerDriver NativeMDNSLockerDriver
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{}
}

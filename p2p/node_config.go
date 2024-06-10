package p2p

// Config is used in NewNode.
type NodeConfig struct {
	netDriver        NativeNetDriver
	mdnsLockerDriver NativeMDNSLockerDriver
	storagePath string
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{
	}
}

func (c *NodeConfig) SetNetDriver(driver NativeNetDriver)         { c.netDriver = driver }
func (c *NodeConfig) SetMDNSLocker(driver NativeMDNSLockerDriver) { c.mdnsLockerDriver = driver }
func (c *NodeConfig) SetStorage(path string) { c.storagePath = path }
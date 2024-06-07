package p2p

//https://github.com/ipfs-shipyard/gomobile-ipfs/pull/139/files

type NativeMDNSLockerDriver interface {
	Lock()
	Unlock()
}

type noopNativeMDNSLockerDriver struct{}

func (*noopNativeMDNSLockerDriver) Lock()   {}
func (*noopNativeMDNSLockerDriver) Unlock() {}
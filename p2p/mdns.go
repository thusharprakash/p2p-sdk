package p2p

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"p2p-sdk/p2p/internals"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	p2p_mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/zeroconf/v2"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

func SetupDiscovery(h host.Host) error {
	mdnslogger, _ := zap.NewDevelopment()
	s := NewMdnsService(mdnslogger, h, DiscoveryServiceTag, &discoveryNotifee{h: h})

	LogToNative("Getting multicast interfaces")
	ifaces, err := GetMulticastInterfaces()
	if err != nil {
		LogToNative("Failed to get multicast interfaces -> "+ err.Error())
		fmt.Println("Failed to get multicast interfaces", err)
		return err
	}

	// If multicast interfaces are found, start the mDNS service.
	if len(ifaces) > 0 {
		LogToNative("Starting mDNS")
		mdnslogger.Info("starting mdns")
		if err := s.Start(); err != nil {
			LogToNative("Failed to start mDNS -> "+ err.Error())
			return nil
		}
	} else {
		LogToNative("No multicast interfaces found")
		mdnslogger.Error("unable to start mdns service, no multicast interfaces found")
	}
	return nil
}

var DiscoveryTimeout = time.Second * 30

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		LogToNative("Discarding self connection to Peer")
		fmt.Println("discarding self connection to Peer")
		return
	}
	go func() {
		var connectContext, cancel = context.WithTimeout(context.Background(), DiscoveryTimeout)
		defer cancel()
		for i := 0; i < 30; i++ {
			fmt.Printf("discovered new peer %s\n", pi)
			LogToNative("Discovered new peer " + pi.ID.String());
			err := n.h.Connect(connectContext, pi)
			if err != nil {
				LogToNative("Error connecting to peer "+ pi.ID.String() + " -> "+ err.Error())
				fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
			} else {
				LogToNative("Connected to peer "+ pi.ID.String())
				fmt.Printf("connected to peer %s\n", pi.ID)
				break
			}
			time.Sleep(time.Duration(rand.Intn(2900)+100) * time.Millisecond)
		}
	}()
}

// NativeMDNSLockerDriver is an interface for locking and unlocking the MDNS service

const (
	MDNSServiceName = p2p_mdns.ServiceName
	mdnsDomain      = "local"
	dnsaddrPrefix   = "dnsaddr="
)

var _ p2p_mdns.Service = (*mdnsService)(nil)

type mdnsService struct {
	logger *zap.Logger

	host        host.Host
	serviceName string
	peerName    string

	// The context is canceled when Close() is called.
	ctx       context.Context
	ctxCancel context.CancelFunc

	resolverWG sync.WaitGroup
	server     *zeroconf.Server

	notifee p2p_mdns.Notifee
}

var _ p2p_mdns.Notifee = (*discoveryHandler)(nil)

type discoveryHandler struct {
	logger *zap.Logger
	ctx    context.Context
	host   host.Host
}

func DiscoveryHandler(ctx context.Context, l *zap.Logger, h host.Host) p2p_mdns.Notifee {
	return &discoveryHandler{
		ctx:    ctx,
		logger: l,
		host:   h,
	}
}

func (dh *discoveryHandler) HandlePeerFound(p peer.AddrInfo) {
	if p.ID == dh.host.ID() {
		dh.logger.Debug("discarding self dialing")
		return
	}

	ctx, cancel := context.WithTimeout(dh.ctx, DiscoveryTimeout)
	defer cancel()

	if err := dh.host.Connect(ctx, p); err != nil {
		dh.logger.Error("failed to connect to peer")
	} else {
		dh.logger.Info("connected to discovered peer")
	}
}

func NewMdnsService(logger *zap.Logger, host host.Host, serviceName string, notifee p2p_mdns.Notifee) *mdnsService {
	if serviceName == "" {
		serviceName = p2p_mdns.ServiceName
	}

	s := &mdnsService{
		logger:      logger,
		host:        host,
		serviceName: serviceName,
		// generate a random string between 32 and 63 characters long
		peerName: randomString(32 + rand.Intn(32)), // nolint:gosec
		notifee:  notifee,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	return s
}

func (s *mdnsService) Start() error {
	if err := s.startServer(); err != nil {
		return err
	}
	s.startResolver(s.ctx)
	return nil
}

func (s *mdnsService) Close() error {
	s.ctxCancel()
	if s.server != nil {
		s.server.Shutdown()
	}
	s.resolverWG.Wait()
	return nil
}

// We don't really care about the IP addresses, but the spec (and various routers / firewalls) require us
// to send A and AAAA records.
func (s *mdnsService) getIPs(addrs []ma.Multiaddr) ([]string, error) {
	var ip4, ip6 string
	for _, addr := range addrs {
		network, hostport, err := manet.DialArgs(addr)
		if err != nil {
			continue
		}
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			continue
		}
		if ip4 == "" && (network == "udp4" || network == "tcp4") {
			ip4 = host
		} else if ip6 == "" && (network == "udp6" || network == "tcp6") {
			ip6 = host
		}
	}
	ips := make([]string, 0, 2)
	if ip4 != "" {
		ips = append(ips, ip4)
	}
	if ip6 != "" {
		ips = append(ips, ip6)
	}
	if len(ips) == 0 {
		LogToNative("Didn't find any IP addresses")
		return nil, errors.New("didn't find any IP addresses")
	}
	return ips, nil
}

func (s *mdnsService) startServer() error {
	interfaceAddrs, err := s.host.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
		ID:    s.host.ID(),
		Addrs: interfaceAddrs,
	})
	if err != nil {
		return err
	}
	var txts []string
	for _, addr := range addrs {
		if manet.IsThinWaist(addr) { // don't announce circuit addresses
			txts = append(txts, dnsaddrPrefix+addr.String())
		}
	}

	ips, err := s.getIPs(addrs)
	if err != nil {
		return err
	}

	// manually get interfaces list
	ifaces, err := GetMulticastInterfaces()
	if err != nil {
		return err
	}
	s.logger.Debug("multicast interfaces found", zap.Int("ifaces", len(ifaces)))
	LogToNative("Multicast interfaces found: "+ fmt.Sprint(len(ifaces)))

	server, err := zeroconf.RegisterProxy(
		s.peerName,
		s.serviceName,
		mdnsDomain,
		4001, // we have to pass in a port number here, but libp2p only uses the TXT records
		s.peerName,
		ips,
		txts,
		ifaces,
	)
	if err != nil {
		return err
	}
	s.server = server
	return nil
}

func (s *mdnsService) startResolver(ctx context.Context) {
	s.resolverWG.Add(2)
	entryChan := make(chan *zeroconf.ServiceEntry, 1000)
	go func() {
		defer s.resolverWG.Done()
		for entry := range entryChan {
			// We only care about the TXT records.
			// Ignore A, AAAA and PTR.
			addrs := make([]ma.Multiaddr, 0, len(entry.Text)) // assume that all TXT records are dnsaddrs
			for _, text := range entry.Text {
				if !strings.HasPrefix(text, dnsaddrPrefix) {
					s.logger.Warn("missing dnsaddr prefix")
					LogToNative("Missing dnsaddr prefix")
					continue
				}
				addr, err := ma.NewMultiaddr(text[len(dnsaddrPrefix):])
				if err != nil {
					s.logger.Warn("failed to parse multiaddr", zap.String("addr", text[len(dnsaddrPrefix):]), zap.Error(err))
					LogToNative("Failed to parse multiaddr -> "+ err.Error())
					continue
				}
				addrs = append(addrs, addr)
			}
			infos, err := peer.AddrInfosFromP2pAddrs(addrs...)
			if err != nil {
				s.logger.Debug("failed to get peer info", zap.Error(err))
				LogToNative("Failed to get peer info -> "+ err.Error())
				continue
			}
			for _, info := range infos {
				go s.notifee.HandlePeerFound(info)
			}
		}
	}()
	go func() {
		// manually get interfaces list
		ifaces, err := internals.GetNetDriver().Interfaces()
		if err != nil {
			s.logger.Error("zeroconf failed to get device interfaces", zap.Error(err))
			LogToNative("Zeroconf failed to get device interfaces -> "+ err.Error())
			return
		}
		// filter Multicast interfaces
		ifaces = filterMulticastInterfaces(ifaces)

		defer s.resolverWG.Done()
		if err := zeroconf.Browse(ctx, s.serviceName, mdnsDomain, entryChan, zeroconf.SelectIfaces(ifaces)); err != nil {
			s.logger.Error("zeroconf browsing failed", zap.Error(err))
			LogToNative("Zeroconf browsing failed -> "+ err.Error())
		}
	}()
}

func GetMulticastInterfaces() ([]net.Interface, error) {
	// manually get interfaces list
	ifaces, err := internals.GetNetDriver().Interfaces()
	if err != nil {
		return nil, err
	}

	// filter Multicast interfaces
	return filterMulticastInterfaces(ifaces), nil
}

func randomString(l int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	s := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		s = append(s, alphabet[rand.Intn(len(alphabet))]) // nolint:gosec
	}
	return string(s)
}

func filterMulticastInterfaces(ifaces []net.Interface) []net.Interface {
	interfaces := []net.Interface{}
	for _, ifi := range ifaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		if (ifi.Flags & net.FlagMulticast) > 0 {
			interfaces = append(interfaces, ifi)
		}
	}

	return interfaces
}

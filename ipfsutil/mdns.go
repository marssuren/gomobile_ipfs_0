package ipfsutil

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	p2p_mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/zeroconf/v2"
	"go.uber.org/zap"
)

const (
	MDNSServiceName = p2p_mdns.ServiceName
	mdnsDomain      = "local"
	dnsaddrPrefix   = "dnsaddr="
)

var _ p2p_mdns.Notifee = (*discoveryHandler)(nil)
var DiscoveryTimeout = time.Second * 30

type discoveryHandler struct {
	logger *zap.Logger
	ctx    context.Context
	host   host.Host
}

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

var _ p2p_mdns.Service = (*mdnsService)(nil)

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

func NewMdnsService(logger *zap.Logger, host host.Host, serviceName string, notifee p2p_mdns.Notifee) p2p_mdns.Service {
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

func randomString(l int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	s := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		s = append(s, alphabet[rand.Intn(len(alphabet))]) // nolint:gosec
	}
	return string(s)
}

func (s *mdnsService) Close() error {
	s.ctxCancel()
	if s.server != nil {
		s.server.Shutdown()
	}
	s.resolverWG.Wait()
	return nil
}

func (s *mdnsService) Start() error {
	s.logger.Info("starting mdns service")
	// 创建服务器实例
	var err error
	s.server, err = zeroconf.Register(s.peerName, s.serviceName, "local.", 4001, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func GetMulticastInterfaces() ([]net.Interface, error) {
	// manually get interfaces list
	ifaces, err := getNetDriver().Interfaces()
	if err != nil {
		return nil, err
	}

	// filter Multicast interfaces
	return filterMulticastInterfaces(ifaces), nil
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

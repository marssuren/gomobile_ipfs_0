package main

import (
	"context"
	"fmt"
	"log"
	"net"

	ipfs_bs "github.com/ipfs/boxo/bootstrap" // IPFS引导节点
	ipfs_config "github.com/ipfs/kubo/config"
	ipfsutil "github.com/marssuren/gomobile_ipfs_0/ipfsutil"
	"go.uber.org/zap"

	p2p_mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns" // mDNS服务发现

	libp2p "github.com/libp2p/go-libp2p" // P2P网络库
)

func main() {

	config := NewNodeConfig()

	// 设置DNS解析器，使用固定的DNS服务器
	var dialer net.Dialer
	net.DefaultResolver = &net.Resolver{
		PreferGo: false, // 不使用Go的DNS解析器
		Dial: func(context context.Context, _, _ string) (net.Conn, error) {
			// 使用硬编码的DNS服务器(84.200.69.80是privacy-friendly的DNS服务器)
			conn, err := dialer.DialContext(context, "udp", "84.200.69.80:53")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	// 创建上下文
	ctx := context.Background()

	repoPath := "./ipfs_repo"

	cfg, err := NewDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	initRepoErr := InitRepo(repoPath, cfg)

	if initRepoErr != nil {
		log.Fatal(initRepoErr)
	}

	r, err := OpenRepo(repoPath)

	// 配置IPFS节点
	ipfscfg := &IpfsConfig{
		HostConfig: &HostConfig{
			Options: []libp2p.Option{},
		},
		RepoMobile: r.mr, // 设置仓库
		ExtraOpts: map[string]bool{
			"pubsub": true, // 默认启用实验性的pubsub功能
			"ipnsps": true, // 默认启用通过pubsub分发IPNS记录
		},
	}

	// 获取仓库配置
	repoCfg, err := r.mr.Config()
	if err != nil {
		panic(err)
	}

	// mDNS处理（多播DNS，用于本地网络发现）
	mdnsLocked := false
	if repoCfg.Discovery.MDNS.Enabled && config.mdnsLockerDriver != nil {
		// 锁定mDNS（避免多个进程同时使用）
		config.mdnsLockerDriver.Lock()
		mdnsLocked = true

		// 暂时禁用mDNS，避免ipfs_mobile.NewNode启动它
		err := r.mr.ApplyPatchs(func(cfg *ipfs_config.Config) error {
			cfg.Discovery.MDNS.Enabled = false
			return nil
		})
		if err != nil {
			fmt.Errorf("unable to ApplyPatchs to disable mDNS: %w", err)
			return
		}
	}

	node, err := NewNode(ctx, ipfscfg)
	if err != nil {
		log.Fatal(err)
		if mdnsLocked {
			config.mdnsLockerDriver.Unlock()
		}
		return
	}

	var mdnsService p2p_mdns.Service = nil
	if mdnsLocked {
		// 恢复mDNS配置
		err := r.mr.ApplyPatchs(func(cfg *ipfs_config.Config) error {
			cfg.Discovery.MDNS.Enabled = true
			return nil
		})
		if err != nil {
			fmt.Errorf("unable to ApplyPatchs to enable mDNS: %w", err)
			return
		}
		// 获取对等节点主机
		h := node.PeerHost()
		mdnslogger, _ := zap.NewDevelopment()
		// 创建发现处理器和mDNS服务
		dh := ipfsutil.DiscoveryHandler(ctx, mdnslogger, h)
		mdnsService = ipfsutil.NewMdnsService(mdnslogger, h, ipfsutil.MDNSServiceName, dh)
		// 启动mDNS服务
		// 获取多播接口
		ifaces, err := ipfsutil.GetMulticastInterfaces()
		if err != nil {
			if mdnsLocked {
				config.mdnsLockerDriver.Unlock()
			}
			fmt.Errorf("unable to GetMulticastInterfaces: %w", err)
			return
		}
		// 如果找到多播接口，启动mDNS服务
		if len(ifaces) > 0 {
			mdnslogger.Info("starting mdns")
			if err := mdnsService.Start(); err != nil {
				if mdnsLocked {
					config.mdnsLockerDriver.Unlock()
				}
				fmt.Errorf("unable to start mdns service: %w", err)
				return
			}
		} else {
			mdnslogger.Error("unable to start mdns service, no multicast interfaces found")
		}
	}

	// 使用默认配置引导节点
	if err := node.IpfsNode.Bootstrap(ipfs_bs.DefaultBootstrapConfig); err != nil {
		log.Printf("failed to bootstrap node: `%s`", err)
	}

}

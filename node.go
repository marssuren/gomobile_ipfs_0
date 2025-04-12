package main

import (
	"context"
	"fmt"
	"sync"

	ipfs_oldcmds "github.com/ipfs/kubo/commands" // IPFS命令接口
	ipfs_core "github.com/ipfs/kubo/core"        // IPFS核心实现
	ipfs_p2p "github.com/ipfs/kubo/core/node/libp2p"
	p2p_host "github.com/libp2p/go-libp2p/core/host"
	p2p_mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	manet "github.com/multiformats/go-multiaddr/net" // 多地址网络接口
)

type IpfsConfig struct {
	HostConfig *HostConfig
	HostOption ipfs_p2p.HostOption

	RoutingConfig *RoutingConfig
	RoutingOption ipfs_p2p.RoutingOption

	RepoMobile *RepoMobile
	ExtraOpts  map[string]bool
}

type Node struct {
	listeners   []manet.Listener // 网络监听器列表
	muListeners sync.Mutex       // 保护listeners的互斥锁
	mdnsLocker  sync.Locker      // mDNS锁，控制mDNS服务的访问
	mdnsLocked  bool             // 标记mDNS是否被锁定
	mdnsService p2p_mdns.Service // mDNS服务，用于本地网络发现

	ipfsMobile *IpfsMobile // 移动平台IPFS节点实例
}

// IpfsMobile是移动平台IPFS节点实现
// 封装了标准IPFS节点并添加移动优化功能
type IpfsMobile struct {
	// 嵌入IPFS核心节点
	*ipfs_core.IpfsNode
	// 引用移动平台仓库
	Repo *RepoMobile

	// 命令上下文，用于HTTP API
	commandCtx ipfs_oldcmds.Context
}

// NewNode根据给定配置创建新的IPFS移动节点
// 这是创建IPFS节点的主要入口点
func NewNode(ctx context.Context, cfg *IpfsConfig) (*IpfsMobile, error) {
	// 填充默认配置值
	if err := cfg.fillDefault(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// 构建IPFS节点配置
	buildcfg := &ipfs_core.BuildCfg{
		Online:                      true,                                                         // 节点处于在线模式
		Permanent:                   false,                                                        // 非永久节点(适合移动设备)
		DisableEncryptedConnections: false,                                                        // 使用加密连接
		Repo:                        cfg.RepoMobile,                                               // 使用移动仓库
		Host:                        NewHostConfigOption(cfg.HostOption, cfg.HostConfig),          // 配置网络主机
		Routing:                     NewRoutingConfigOption(cfg.RoutingOption, cfg.RoutingConfig), // 配置路由
		ExtraOpts:                   cfg.ExtraOpts,                                                // 设置额外选项(如pubsub)
	}

	// 创建IPFS核心节点
	inode, err := ipfs_core.NewNode(ctx, buildcfg)
	if err != nil {
		// 注释掉了解锁仓库的代码
		// unlockRepo(repoPath)
		return nil, fmt.Errorf("failed to init ipfs node: %s", err)
	}

	// 创建命令上下文
	// 注释表明这可能不是初始化的最佳方式
	cctx := ipfs_oldcmds.Context{
		ConfigRoot: cfg.RepoMobile.Path(),  // 配置根路径
		ReqLog:     &ipfs_oldcmds.ReqLog{}, // 请求日志
		ConstructNode: func() (*ipfs_core.IpfsNode, error) { // 节点构造函数
			return inode, nil
		},
	}

	// 返回创建的移动IPFS节点
	return &IpfsMobile{
		commandCtx: cctx,           // 命令上下文
		IpfsNode:   inode,          // IPFS核心节点
		Repo:       cfg.RepoMobile, // 仓库引用
	}, nil
}

// fillDefault为配置填充默认值
// 确保配置对象包含所有必需的字段
func (c *IpfsConfig) fillDefault() error {
	// 仓库是必需的，不能为空
	if c.RepoMobile == nil {
		return fmt.Errorf("repo cannot be nil")
	}

	// 如果额外选项为空，创建空映射
	if c.ExtraOpts == nil {
		c.ExtraOpts = make(map[string]bool)
	}

	// 默认使用DHT(分布式哈希表)作为路由选项
	if c.RoutingOption == nil {
		c.RoutingOption = ipfs_p2p.DHTOption
	}

	// 如果没有路由配置，创建默认配置
	if c.RoutingConfig == nil {
		c.RoutingConfig = &RoutingConfig{}
	}

	// 默认使用标准主机选项
	if c.HostOption == nil {
		c.HostOption = ipfs_p2p.DefaultHostOption
	}

	// 如果没有主机配置，创建默认配置
	if c.HostConfig == nil {
		c.HostConfig = &HostConfig{}
	}

	return nil
}

// PeerHost返回节点的P2P网络主机
// 允许访问底层网络功能
func (im *IpfsMobile) PeerHost() p2p_host.Host {
	return im.IpfsNode.PeerHost
}

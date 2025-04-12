package main

import (
	"fmt"

	// IPFS数据存储接口
	ipfs_p2p "github.com/ipfs/kubo/core/node/libp2p" // libp2p记录验证
	p2p_host "github.com/libp2p/go-libp2p/core/host" // libp2p主机接口
	// 对等节点标识
	p2p_routing "github.com/libp2p/go-libp2p/core/routing" // 内容路由接口
)

// RoutingConfigFunc定义配置路由系统的函数类型
// 接收主机和路由实例，可对路由进行配置，返回可能的错误
type RoutingConfigFunc func(p2p_host.Host, p2p_routing.Routing) error

// RoutingConfig定义路由系统的配置选项
// 与Host配置结构相似，但专注于路由系统
type RoutingConfig struct {
	ConfigFunc RoutingConfigFunc // 路由配置函数
}

// NewRoutingConfigOption创建新的IPFS路由配置选项
// 将自定义路由配置与IPFS标准路由系统集成
// 参数:
//
//	ro: 基础IPFS路由选项
//	rc: 自定义路由配置
//
// 返回:
//
//	集成了自定义配置的IPFS路由选项函数
func NewRoutingConfigOption(ro ipfs_p2p.RoutingOption, rc *RoutingConfig) ipfs_p2p.RoutingOption {
	return func(args ipfs_p2p.RoutingOptionArgs) (p2p_routing.Routing, error) {
		// 使用基础选项创建路由系统
		routing, err := ro(args)
		if err != nil {
			return nil, err
		}

		// 如果提供了配置函数，应用它
		if rc.ConfigFunc != nil {
			if err := rc.ConfigFunc(args.Host, routing); err != nil {
				return nil, fmt.Errorf("failed to config routing: %w", err)
			}
		}

		// 返回配置好的路由系统
		return routing, nil
	}
}

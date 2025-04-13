package core

import (
	// libp2p核心库
	"fmt"

	ipfs_p2p "github.com/ipfs/kubo/core/node/libp2p"
	p2p "github.com/libp2p/go-libp2p"                       // libp2p网络库主包
	p2p_host "github.com/libp2p/go-libp2p/core/host"        // 网络主机接口
	p2p_peer "github.com/libp2p/go-libp2p/core/peer"        // 对等节点标识
	p2p_pstore "github.com/libp2p/go-libp2p/core/peerstore" // 对等节点存储
)

// HostConfigFunc定义一个函数类型，用于配置主机
// 它接收创建好的主机实例，可以对其进行配置，并返回可能的错误
type HostConfigFunc func(p2p_host.Host) error

// HostConfig定义主机的配置选项
// @TODO: 注释表明这里计划添加更多移动平台特定的选项
type HostConfig struct {
	// 主机初始化后调用的配置函数
	ConfigFunc HostConfigFunc

	// libp2p网络选项列表，可以包含传输协议、安全选项等
	Options []p2p.Option
}

// NewHostConfigOption创建一个新的IPFS主机配置选项
// 这个函数接合了IPFS的主机选项系统和我们自定义的HostConfig
// 参数:
//
//	hopt: 基础IPFS主机选项
//	cfg: 自定义主机配置
//
// 返回:
//
//	一个新的IPFS主机选项函数，集成了自定义配置
func NewHostConfigOption(hopt ipfs_p2p.HostOption, cfg *HostConfig) ipfs_p2p.HostOption {
	// 返回符合IPFS主机选项接口的函数
	return func(id p2p_peer.ID, ps p2p_pstore.Peerstore, options ...p2p.Option) (p2p_host.Host, error) {
		// 添加自定义P2P选项
		if cfg.Options != nil {
			options = append(options, cfg.Options...)
		}

		// 使用基础选项创建主机
		host, err := hopt(id, ps, options...)
		if err != nil {
			return nil, err
		}

		// 如果提供了配置函数，应用它
		if cfg.ConfigFunc != nil {
			// 应用自定义主机配置
			if err := cfg.ConfigFunc(host); err != nil {
				return nil, fmt.Errorf("unable to apply host config: %w", err)
			}
		}

		// 返回配置好的主机
		return host, nil
	}
}

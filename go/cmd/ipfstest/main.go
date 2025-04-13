package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/marssuren/gomobile_ipfs_0/go/bind/core"
)

func main() {
	// 创建一个上下文用于取消操作
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理系统信号，以便优雅地关闭
	setupSignalHandler(cancel)

	// 定义仓库路径
	repoPath := filepath.Join(".", "ipfs_repo_test")

	// 打印欢迎信息
	fmt.Println("===== IPFS节点测试程序 =====")
	fmt.Printf("仓库路径: %s\n", repoPath)

	// 初始化或打开仓库
	repo, err := initOrOpenRepo(repoPath)
	if err != nil {
		fmt.Printf("初始化/打开仓库失败: %s\n", err)
		os.Exit(1)
	}

	// 注意：这里不需要 defer repo.Close()，因为 Repo 结构体没有直接的 Close 方法
	// 关闭会通过 ipfsMobile 完成

	// 创建IPFS节点配置
	ipfsConfig := &core.IpfsConfig{
		RepoMobile: repo.Mobile(), // 使用 Mobile() 方法访问 RepoMobile
		ExtraOpts: map[string]bool{
			"pubsub": true, // 启用pubsub功能
			"ipnsps": true, // 启用IPNS over pubsub
		},
	}

	// 创建并启动IPFS节点
	fmt.Println("正在启动IPFS节点...")
	ipfsMobile, err := core.NewNode(ctx, ipfsConfig)
	if err != nil {
		fmt.Printf("启动节点失败: %s\n", err)
		os.Exit(1)
	}

	// 获取节点ID
	id := ipfsMobile.PeerHost().ID()
	fmt.Printf("节点ID: %s\n", id.String())

	// 启动引导过程
	fmt.Println("正在连接到IPFS网络...")
	// 如果有bootstrap方法，可以在这里调用

	// 保持程序运行，直到接收到取消信号
	fmt.Println("节点已启动并运行。按Ctrl+C停止...")
	<-ctx.Done()

	// 关闭节点
	fmt.Println("\n正在关闭IPFS节点...")
	if err := ipfsMobile.IpfsNode.Close(); err != nil {
		fmt.Printf("关闭节点时出错: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("节点已成功关闭")
}

// 初始化或打开IPFS仓库
func initOrOpenRepo(repoPath string) (*core.Repo, error) {
	// 检查仓库是否已经存在
	if _, err := os.Stat(filepath.Join(repoPath, "config")); os.IsNotExist(err) {
		fmt.Println("仓库不存在，正在初始化...")

		// 创建默认配置
		cfg, err := core.NewDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("创建默认配置失败: %w", err)
		}

		// 初始化仓库
		if err := core.InitRepo(repoPath, cfg); err != nil {
			return nil, fmt.Errorf("初始化仓库失败: %w", err)
		}

		fmt.Println("仓库初始化成功")
	} else {
		fmt.Println("找到现有仓库，正在打开...")
	}

	// 打开仓库
	return core.OpenRepo(repoPath)
}

// 设置信号处理器来处理Ctrl+C和其他终止信号
func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n接收到终止信号")
		cancel()

		// 如果5秒后程序还未退出，强制退出
		go func() {
			time.Sleep(5 * time.Second)
			fmt.Println("强制退出")
			os.Exit(1)
		}()
	}()
}

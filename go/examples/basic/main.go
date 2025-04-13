package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/core/coreapi"
	coreiface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/marssuren/gomobile_ipfs_0/go/bind/core"
)

func main() {
	// 创建一个上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 定义仓库路径
	repoPath := filepath.Join(".", "ipfs_repo_example")

	// 打印欢迎信息
	fmt.Println("===== IPFS基本示例程序 =====")
	fmt.Println("此示例展示如何添加和获取内容")
	fmt.Printf("仓库路径: %s\n", repoPath)

	// 初始化或打开仓库
	repo, err := initOrOpenRepo(repoPath)
	if err != nil {
		fmt.Printf("初始化/打开仓库失败: %s\n", err)
		os.Exit(1)
	}

	// 创建IPFS节点配置
	ipfsConfig := &core.IpfsConfig{
		RepoMobile: repo.Mobile(),
		ExtraOpts: map[string]bool{
			"pubsub": true,
			"ipnsps": true,
		},
	}

	// 创建并启动IPFS节点
	fmt.Println("正在启动IPFS节点...")
	ipfsMobile, err := core.NewNode(ctx, ipfsConfig)
	if err != nil {
		fmt.Printf("启动节点失败: %s\n", err)
		os.Exit(1)
	}
	defer ipfsMobile.IpfsNode.Close()

	// 获取节点ID
	id := ipfsMobile.PeerHost().ID()
	fmt.Printf("节点ID: %s\n", id.String())

	// 创建IPFS API
	api, err := coreapi.NewCoreAPI(ipfsMobile.IpfsNode)
	if err != nil {
		fmt.Printf("创建API失败: %s\n", err)
		os.Exit(1)
	}

	// 示例内容
	content := "Hello, IPFS!"

	// 添加内容
	fmt.Println("\n1. 添加内容到IPFS")
	cid, err := addContent(ctx, api, content)
	if err != nil {
		fmt.Printf("添加内容失败: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("内容已添加，CID: %s\n", cid)

	// 获取内容
	fmt.Println("\n2. 从IPFS获取内容")
	retrievedContent, err := getContent(ctx, api, cid)
	if err != nil {
		fmt.Printf("获取内容失败: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("检索的内容: %s\n", retrievedContent)

	// 验证内容
	if content == retrievedContent {
		fmt.Println("\n✓ 内容匹配！示例成功。")
	} else {
		fmt.Println("\n✗ 内容不匹配！示例失败。")
	}
}

// 初始化或打开IPFS仓库
func initOrOpenRepo(repoPath string) (*core.Repo, error) {
	if _, err := os.Stat(filepath.Join(repoPath, "config")); os.IsNotExist(err) {
		fmt.Println("仓库不存在，正在初始化...")

		cfg, err := core.NewDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("创建默认配置失败: %w", err)
		}

		if err := core.InitRepo(repoPath, cfg); err != nil {
			return nil, fmt.Errorf("初始化仓库失败: %w", err)
		}

		fmt.Println("仓库初始化成功")
	} else {
		fmt.Println("找到现有仓库，正在打开...")
	}

	return core.OpenRepo(repoPath)
}

// 添加内容到IPFS
func addContent(ctx context.Context, api coreiface.CoreAPI, content string) (string, error) {
	// 创建一个内存中的文件
	r := strings.NewReader(content)
	fileNode := files.NewReaderFile(r)

	// 添加到IPFS
	path, err := api.Unixfs().Add(ctx, fileNode, options.Unixfs.Pin(true))
	if err != nil {
		return "", err
	}

	return path.Cid().String(), nil
}

// 从IPFS获取内容
func getContent(ctx context.Context, api coreiface.CoreAPI, cid string) (string, error) {
	// 解析路径
	path, err := coreiface.ParsePath("/ipfs/" + cid)
	if err != nil {
		return "", err
	}

	// 获取内容
	node, err := api.Unixfs().Get(ctx, path)
	if err != nil {
		return "", err
	}

	// 读取内容
	f, ok := node.(files.File)
	if !ok {
		return "", fmt.Errorf("not a file")
	}

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

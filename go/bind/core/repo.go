package core

import (
	"path/filepath"
	"sync"

	ipfs_config "github.com/ipfs/kubo/config"
	ipfs_loader "github.com/ipfs/kubo/plugin/loader" // IPFS插件加载器
	ipfs_repo "github.com/ipfs/kubo/repo"            // IPFS仓库接口
	ipfs_fsrepo "github.com/ipfs/kubo/repo/fsrepo"   // 基于文件系统的IPFS仓库实现
)

var (
	// 全局变量，用于插件管理
	muPlugins sync.Mutex                // 保护plugins变量的互斥锁
	plugins   *ipfs_loader.PluginLoader // 全局插件加载器实例
)

// Repo 结构体包装了移动平台的IPFS仓库
type Repo struct {
	mr *RepoMobile // 指向移动平台IPFS仓库的指针
}

// RepoConfigPatch定义一个函数类型，用于修改IPFS配置
// 每个补丁函数接收一个配置指针并可以修改它，返回可能的错误
type RepoConfigPatch func(cfg *ipfs_config.Config) (err error)

// RepoMobile是标准IPFS仓库的移动平台封装
// 它添加了路径信息和配置补丁功能
type RepoMobile struct {
	// 嵌入标准IPFS仓库接口，继承其所有方法
	ipfs_repo.Repo
	path string
}

// 添加方法实现接口要求
func (r *RepoMobile) Path() string {
	return r.path
}

// InitRepo 在指定路径初始化IPFS仓库
func InitRepo(path string, cfg *Config) error {
	// 加载插件，确保初始化仓库前插件系统已就绪
	if _, err := loadPlugins(path); err != nil {
		return err
	}

	// 使用配置初始化仓库
	return ipfs_fsrepo.Init(path, cfg.getConfig())
}

// OpenRepo 打开现有的IPFS仓库
func OpenRepo(path string) (*Repo, error) {
	// 加载插件，确保打开仓库前插件系统已就绪
	if _, err := loadPlugins(path); err != nil {
		return nil, err
	}

	// 打开标准IPFS仓库
	irepo, err := ipfs_fsrepo.Open(path)
	if err != nil {
		return nil, err
	}

	// 创建移动平台适用的仓库包装
	mRepo := NewRepoMobile(path, irepo)
	return &Repo{mRepo}, nil
}

// loadPlugins 加载IPFS插件系统
func loadPlugins(repoPath string) (*ipfs_loader.PluginLoader, error) {
	// 加锁确保多线程安全
	muPlugins.Lock()
	defer muPlugins.Unlock() // 确保函数退出时解锁

	// 如果插件已加载，直接返回现有实例（单例模式）
	if plugins != nil {
		return plugins, nil
	}

	// 构建插件目录路径
	// 默认IPFS插件存放在仓库的"plugins"子目录
	pluginpath := filepath.Join(repoPath, "plugins")

	// 创建新的插件加载器
	lp, err := ipfs_loader.NewPluginLoader(pluginpath)
	if err != nil {
		return nil, err
	}

	// 初始化插件系统
	// 这会查找和加载所有可用插件的元数据
	if err = lp.Initialize(); err != nil {
		return nil, err
	}

	// 注入插件
	// 这会将插件实际集成到IPFS系统中，使其功能可用
	if err = lp.Inject(); err != nil {
		return nil, err
	}

	// 保存全局实例并返回
	plugins = lp
	return lp, nil
}

// NewRepoMobile创建一个新的移动平台仓库实例
// 参数:
//
//	path: 仓库在文件系统中的路径
//	repo: 底层IPFS仓库实现
//
// 返回:
//
//	移动平台仓库实例
func NewRepoMobile(path string, repo ipfs_repo.Repo) *RepoMobile {
	return &RepoMobile{
		Repo: repo, // 存储底层仓库实现
		path: path, // 保存仓库路径
	}
}

// ApplyPatchs应用一系列配置补丁到仓库配置
// 这允许以可组合的方式修改IPFS配置
// 参数:
//
//	patchs: 要应用的配置补丁函数变参
//
// 返回:
//
//	可能的错误
func (mr *RepoMobile) ApplyPatchs(patchs ...RepoConfigPatch) error {
	// 获取当前配置
	cfg, err := mr.Config()
	if err != nil {
		return err
	}

	// 使用链式补丁函数应用所有补丁
	if err := ChainIpfsConfigPatch(patchs...)(cfg); err != nil {
		return err
	}

	// 将修改后的配置保存回仓库
	return mr.SetConfig(cfg)
}

// ChainIpfsConfigPatch将多个配置补丁函数合并为一个
// 这是函数式编程中的组合模式
// 参数:
//
//	patchs: 要链接的配置补丁函数变参
//
// 返回:
//
//	合并后的配置补丁函数
func ChainIpfsConfigPatch(patchs ...RepoConfigPatch) RepoConfigPatch {
	// 返回一个新函数，它将依次应用所有补丁
	return func(cfg *ipfs_config.Config) (err error) {
		// 遍历所有补丁函数
		for _, patch := range patchs {
			// 跳过空补丁
			if patch == nil {
				continue // skip empty patch
			}

			// 应用当前补丁，如果出错则返回
			if err = patch(cfg); err != nil {
				return
			}
		}
		// 所有补丁应用成功
		return
	}
}

// Mobile 返回底层的 RepoMobile 实例
func (r *Repo) Mobile() *RepoMobile {
	return r.mr
}

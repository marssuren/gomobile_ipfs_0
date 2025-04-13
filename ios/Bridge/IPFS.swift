import Foundation
import Ipfs

/**
 * IPFS 是对 Go 实现的 IPFS 节点的 Swift 包装器
 * 为 iOS 应用程序提供了简单易用的接口
 */
@objc public class IPFS: NSObject {
    private var node: CoreNode?
    private var repo: CoreRepo?
    private let repoPath: String
    
    /**
     * 创建新的 IPFS 实例
     * @param repoPath 仓库路径
     */
    @objc public init(repoPath: String) {
        self.repoPath = repoPath
        
        super.init()
        
        // 确保存储目录存在
        let fileManager = FileManager.default
        if !fileManager.fileExists(atPath: repoPath) {
            try? fileManager.createDirectory(atPath: repoPath, withIntermediateDirectories: true)
        }
    }
    
    /**
     * 启动 IPFS 节点
     * @throws 如果启动失败
     */
    @objc public func start() throws {
        // 如果节点已经运行，不做任何操作
        if isStarted() {
            return
        }
        
        // 打开或初始化仓库
        if repo == nil {
            do {
                repo = try CoreOpenRepo(repoPath)
            } catch {
                // 仓库不存在，需要初始化
                let cfg = CoreNewDefaultConfig()
                try CoreInitRepo(repoPath, cfg)
                repo = try CoreOpenRepo(repoPath)
            }
        }
        
        // 创建节点配置
        let config = CoreNewNodeConfig()
        
        // 创建和启动节点
        node = try CoreNewNode(repo!, config)
    }
    
    /**
     * 检查节点是否已启动
     * @return 如果节点已启动返回 true
     */
    @objc public func isStarted() -> Bool {
        return node != nil
    }
    
    /**
     * 关闭 IPFS 节点
     * @throws 如果关闭失败
     */
    @objc public func stop() throws {
        if let node = node {
            try node.close()
            self.node = nil
        }
        
        if let repo = repo {
            try repo.close()
            self.repo = nil
        }
    }
    
    /**
     * 获取节点 ID
     * @return 节点 ID 字符串
     * @throws 如果获取失败
     */
    @objc public func getNodeID() throws -> String {
        if !isStarted() {
            throw NSError(domain: "IPFSError", code: 1, userInfo: [NSLocalizedDescriptionKey: "IPFS node is not started"])
        }
        
        // 这里需要根据实际 API 调整
        // 在原始实现中，可能通过 HTTP API 调用 /id 接口获取
        return "节点 ID 获取功能待实现"
    }
} 
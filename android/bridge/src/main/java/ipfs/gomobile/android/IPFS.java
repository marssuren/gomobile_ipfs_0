package ipfs.gomobile.android;

import org.ipfs.gomobile.core.Core;
import org.ipfs.gomobile.core.Node;
import org.ipfs.gomobile.core.NodeConfig;
import org.ipfs.gomobile.core.Repo;
import org.ipfs.gomobile.core.RepoConfig;

import java.io.File;

/**
 * IPFS 是对 Go 实现的 IPFS 节点的 Java 包装器
 * 为 Android 应用程序提供了简单易用的接口
 */
public class IPFS {
    private Node node;
    private Repo repo;
    private final String repoPath;

    /**
     * 创建新的 IPFS 实例
     * @param repoPath 仓库路径
     * @throws Exception 如果初始化失败
     */
    public IPFS(String repoPath) throws Exception {
        this.repoPath = repoPath;
        
        // 确保存储目录存在
        File dir = new File(repoPath);
        if (!dir.exists()) {
            dir.mkdirs();
        }
    }

    /**
     * 启动 IPFS 节点
     * @throws Exception 如果启动失败
     */
    public void start() throws Exception {
        // 如果节点已经运行，不做任何操作
        if (isStarted()) {
            return;
        }

        // 打开或初始化仓库
        if (repo == null) {
            try {
                repo = Core.openRepo(repoPath);
            } catch (Exception e) {
                // 仓库不存在，需要初始化
                RepoConfig cfg = Core.newDefaultConfig();
                Core.initRepo(repoPath, cfg);
                repo = Core.openRepo(repoPath);
            }
        }

        // 创建节点配置
        NodeConfig config = Core.newNodeConfig();
        
        // 创建和启动节点
        node = Core.newNode(repo, config);
    }

    /**
     * 检查节点是否已启动
     * @return 如果节点已启动返回 true
     */
    public boolean isStarted() {
        return node != null;
    }

    /**
     * 关闭 IPFS 节点
     * @throws Exception 如果关闭失败
     */
    public void stop() throws Exception {
        if (node != null) {
            node.close();
            node = null;
        }
        
        if (repo != null) {
            repo.close();
            repo = null;
        }
    }
    
    /**
     * 获取节点 ID
     * @return 节点 ID 字符串
     * @throws Exception 如果获取失败
     */
    public String getNodeID() throws Exception {
        if (!isStarted()) {
            throw new IllegalStateException("IPFS node is not started");
        }
        
        // 这里需要根据实际 API 调整
        // 在原始实现中，可能通过 HTTP API 调用 /id 接口获取
        return "节点 ID 获取功能待实现";
    }
} 
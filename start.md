# go run main.go
Type 'dlv help' for list of commands.
 20:10:38 =============================================
 20:10:38   迷你区块浏览器与 ERC-20 监听服务 启动中...
 20:10:38 =============================================
 20:10:38 ✅ 配置加载完成
 20:10:38    - 当前网络: sepolia
 20:10:38    - 节点 URL: wss://ethereum-sepolia.publicnode.com
 20:10:38    - ChainID: 11155111
 20:10:38    - ERC20 合约: <合约地址>
 20:10:38 🔗 正在连接以太坊节点...
 20:10:39 ✅ 以太坊节点连接成功
 20:10:39 📦 初始化 SQLite 交易历史存储...
 20:10:39 📦 [TxHistoryStore] SQLite 存储初始化成功，路径: transactions.db
 20:10:39 ✅ SQLite 交易历史存储初始化完成
 20:10:39 🔐 初始化钱包...
 20:10:39 ✅ 钱包初始化成功
 20:10:39 🔧 初始化服务组件...
 20:10:39 🔧 [BlockService] 初始化
 20:10:39 🔧 [TxService] 初始化
 20:10:39 🔧 [EventService] 初始化，合约地址: <合约地址>
 20:10:39 📄 [EventService] ABI 文件加载成功
 20:10:39 ✅ [EventService] ABI 解析成功
 20:10:39 ✅ 交易发送服务初始化完成
 20:10:39 🔧 初始化合约服务...
 20:10:39 ✅ 合约服务初始化完成
 20:10:39 🔧 初始化 ERC20 服务...
 20:10:39 ✅ ERC20 服务初始化完成
 20:10:39 ✅ 服务组件初始化完成
 20:10:39 🔐 签名钱包已配置，全功能模式运行
 20:10:39 👂 启动 ERC20 Transfer 事件监听...
 20:10:39 🌐 启动 HTTP API 服务器 (端口: 8080)...
 20:10:39 👂 [EventService] 开始订阅 ERC20 Transfer 事件...
 20:10:39 ✅ HTTP API 服务器启动完成
 20:10:39 =============================================
 20:10:39   🚀 服务已就绪，等待请求...
 20:10:39 =============================================
 20:10:39 🌐 [HTTP] 服务器启动中，监听地址: :8080
 20:10:39 🗺️  [HTTP] 已注册路由:
 20:10:39    - GET /api/block/{id}
 20:10:39    - GET /api/tx/{hash}
 20:10:39    - GET /api/events
 20:10:39    - POST /api/tx/send
 20:10:39    - GET /api/tx/history
 20:10:39    - GET /api/tx/detail
 20:10:39    - POST /api/contract/view
 20:10:39    - POST /api/contract/call
 20:10:39    - GET /api/token/info
 20:10:39    - GET /api/token/balance
 20:10:39    - POST /api/token/transfer
 20:10:39    - POST /api/token/mint
 20:10:39    - POST /api/token/deploy
 20:10:40 ✅ [EventService] 事件订阅成功，监听合约: <合约地址>

# 第4期：实战项目——迷你区块浏览器与 ERC-20 监听服务

本讲义聚焦项目实战部分，帮助你将前面学到的 Go-Eth Client 能力整合到一个可运行的小项目中。

## 1. 项目目标

实现一个基于 Go-Eth Client 的**迷你区块浏览器/监听服务**，具备以下能力：

- 查询**区块 / 交易 / 地址**的基础信息  
- 实时监听指定 ERC-20 合约的 `Transfer` 事件  
- 提供简单的 HTTP API 或 CLI 接口对外查询  
- 为后续扩展（多合约、多网络、持久化存储）预留空间  

## 2. 功能范围设计

### 2.1 必做功能（MVP）

1. **区块查询**  
   - 根据区块号或哈希，返回：
     - 区块号、哈希、父哈希、时间戳  
     - 交易数量、Gas 使用情况等简要信息  

2. **交易查询**  
   - 根据交易哈希，返回：
     - from / to / value / gas / input data  
     - 回执中的 status / gasUsed / logs 数量等  

3. **ERC-20 转账监听**  
   - 背景协程订阅指定 ERC-20 合约的 `Transfer` 事件  
   - 将最近 N 条事件保存在内存中（如环形队列）  
   - 提供接口查看最近事件列表  

### 2.2 选做功能（进阶）

- 地址视图：
  - 查看某地址最近参与的交易列表  
  - 查看地址当前 ETH 余额（后续可扩展到代币余额）  
- 历史扫描：
  - 程序启动时，从指定高度回放扫描一定区间的历史区块/事件  
- 持久化存储：
  - 将事件写入 SQLite / PostgreSQL，并支持分页查询  

## 3. 推荐项目结构

在 `lesson-04/examples/09-project` 中，我们提供了一个最小可行版本（MVP）的实现：

```
examples/09-project/
├── go.mod
└── main.go                # 单文件实现，包含所有功能
```

当前实现将所有功能集中在 `main.go` 中，包括：
- 事件存储（`EventStore` 结构体，使用内存环形缓冲区）
- ERC-20 Transfer 事件订阅与解码
- HTTP API 接口（`GET /events`）

**扩展建议**：在实际工程中可以根据需求进行拆分，采用分层结构：

```
examples/project/
├── go.mod
├── main.go                # 程序入口
├── config/                # 配置（RPC 地址、监听合约等）
│   └── config.go
├── client/                # Go-Eth Client 封装
│   └── eth_client.go
├── service/               # 业务逻辑：区块/交易/事件
│   ├── block_service.go
│   ├── tx_service.go
│   └── event_service.go
├── api/                   # 对外 API（HTTP or CLI）
│   ├── http_server.go
│   └── handlers.go
└── store/                 # 存储层（内存 / DB）
    └── memory_store.go
```

核心思想是**分层清晰**：Client → Service → API / Store。当前 MVP 版本为了简化，将所有功能集中在一个文件中，便于理解核心流程。

## 4. 核心流程设计

### 4.1 启动流程

1. 读取配置（环境变量）：
   - `ETH_WS_URL` 或 `ETH_RPC_URL`（优先使用 WebSocket URL）  
   - `ERC20_CONTRACT`：监听的 ERC-20 合约地址  
2. 初始化 Go-Eth Client 和 ABI 解析器  
3. 创建事件存储（`EventStore`，限制最多保存 100 条事件）  
4. 启动事件监听协程（`subscribeTransferEvents`，订阅 ERC-20 Transfer 事件）  
5. 启动 HTTP 服务器（监听 `:8080`，提供 `GET /events` 接口）  
6. 捕获系统信号（SIGINT/SIGTERM）实现优雅退出  

### 4.2 事件监听流程（ERC-20 Transfer）

1. 根据配置构建 `ethereum.FilterQuery`：  
   - `Addresses`：合约地址列表（从环境变量 `ERC20_CONTRACT` 读取）  
   - 当前实现暂未设置 `Topics`，会监听该合约的所有事件（可扩展为只监听 Transfer 事件）  
2. 使用 `client.SubscribeFilterLogs(ctx, query, logsCh)` 订阅日志  
3. 在 goroutine 中循环读取日志：
   - 检查 `log.Topics` 长度，跳过无效日志  
   - 使用 ABI 解码事件：`UnpackIntoInterface` 解码 Data 部分（`value`），从 Topics[1] 和 Topics[2] 解析 `from` 和 `to` 地址  
   - 构造 `TransferEvent` 结构体并存入 `EventStore`（使用 `sync.RWMutex` 保证并发安全）  
   - 当存储超过限制（100 条）时，自动丢弃最旧的事件（环形缓冲区）  
4. 当前实现中，订阅错误会导致 goroutine 退出；实际生产环境应实现重连机制（参考 `examples/07-reconnect-strategy/main.go`）  

### 4.3 查询接口设计

以 HTTP API 为例，可设计如下路由（可根据需要简化）：

- `GET /api/block/{number_or_hash}`  
- `GET /api/tx/{hash}`  
- `GET /api/events`（可附带 query 参数过滤合约或地址）  

每个 handler 调用对应的 service 层函数，再由 service 调用 client 进行 RPC 请求或访问 store。

## 5. 关键技术点讲解建议

### 5.1 高效处理日志事件

- 为事件定义结构体（`TransferEvent`），包含：
  - `BlockNumber`、`TxHash`、`From`、`To`、`Value`、`Timestamp` 等字段  
  - 使用 JSON tag 便于序列化输出  
- 在存储层（`EventStore`）维护：
  - 最近 N 条事件的切片（示例中限制为 100 条）  
  - 使用 `sync.RWMutex` 保证并发安全（订阅 goroutine 写入，HTTP handler 读取）  
  - 当超过限制时，删除最旧的事件（`events = events[1:]`）  
  - 按地址索引的 map（选做，当前实现未包含）  

### 5.2 错误与重连处理

- 讲解典型错误：
  - 网络错误、订阅被服务器关闭、RPC 限流等  
- 给出简单重连策略：
  - 出错后等待几秒重试  
  - 使用计数器或日志记录重试次数  

### 5.3 配置与环境

- 使用环境变量控制：
  - RPC 地址  
  - 监听合约列表  
  - 起始高度或历史扫描范围  
- 可选：增加配置文件（YAML/JSON）并用 Viper 等库读取（与第3期 Gin 课程形成呼应）  

## 6. 演示与验证建议

### 6.1 本地链演示流程

1. 启动本地开发链（`geth --dev` 或 `anvil`）  
   - 确保启用 WebSocket RPC（例如 `--ws` 或 `--ws.port 8546`）  
2. 部署一个简单 ERC-20 合约（可使用预先准备好的脚本或 Remix）  
3. 配置环境变量：
   ```bash
   export ETH_WS_URL=ws://127.0.0.1:8546
   export ERC20_CONTRACT=<部署的合约地址>
   ```
4. 运行项目：`go run main.go`  
5. 在本地链上发起几笔 `Transfer` 交易（可通过 Remix、MetaMask 或脚本）  
6. 访问 `http://localhost:8080/events` 查看最近的事件记录（JSON 格式）  

### 6.2 测试网演示流程（选做）

1. 选择公共测试网（如 Sepolia）  
2. 获取测试网的 WebSocket RPC 地址（如 Infura、Alchemy 等服务提供商的 WebSocket URL）  
3. 选择一个活跃的 ERC-20 合约地址（例如 USDC 在 Sepolia 测试网的地址）  
4. 配置环境变量：
   ```bash
   export ETH_WS_URL=wss://sepolia.infura.io/ws/v3/<your-project-id>
   export ERC20_CONTRACT=<ERC-20合约地址>
   ```
5. 运行项目并观察在一段时间内捕获到的真实转账事件  
6. 注意：公共测试网事件频率可能较低，需要等待一段时间才能看到结果  

## 7. 练习与扩展任务

你可以在完成基础功能后继续练习：

1. **练习 1：为地址增加过滤参数**  
   - 在 `/api/events` 增加 `address` 参数，只返回涉及该地址的事件  

2. **练习 2：增加分页能力**  
   - 为事件查询添加 `page` 与 `page_size` 参数  
   - 修改 `EventStore.List()` 方法或 HTTP handler 支持分页返回  

3. **练习 3：增加简单的 Web 页面**  
   - 使用任意前端框架或纯 HTML，展示最近区块与事件列表  
   - 可以添加一个 `GET /` 路由返回 HTML 页面，或使用模板引擎渲染
4. **练习 4：实现重连机制**  
   - 参考 `examples/07-reconnect-strategy/main.go`，为事件订阅添加自动重连功能  
   - 当订阅失败时，等待一段时间后重新建立连接和订阅
5. **练习 5：增加区块和交易查询接口**  
   - 添加 `GET /api/block/:number` 和 `GET /api/tx/:hash` 接口  
   - 复用前面示例中的区块和交易查询逻辑  


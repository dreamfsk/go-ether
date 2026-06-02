# go-ether

一个基于 Go 语言和 go-ethereum 库构建的迷你区块浏览器和 ERC-20 监控服务。提供完整的区块链数据查询、交易发送、合约交互和代币管理功能。

## 📋 功能特性

### 核心功能
- **区块查询**：支持通过区块号或哈希查询区块详情
- **交易查询**：查询交易详情、输入数据和回执信息
- **ERC-20 事件监听**：实时监听 Transfer 事件，自动保存到内存和数据库
- **交易发送**：支持 ETH 转账，自动处理 Gas 估算和交易签名
- **合约交互**：支持调用合约视图方法和发送合约交易
- **代币管理**：查询代币信息、余额和进行代币转账

### 架构特性
- **分层架构**：Client → Service → API / Store
- **多网络支持**：支持 Sepolia 测试网和本地测试链
- **持久化存储**：SQLite 存储交易历史和事件
- **优雅关闭**：支持 SIGINT/SIGTERM 信号处理
- **实时监控**：WebSocket 订阅 ERC-20 事件

## 🏗️ 架构设计

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP API Layer                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐ ┌───────────────┐    │
│  │ /api/   │ │ /api/tx │ │ /api/token  │ │ /api/contract │    │
│  │ block   │ │ /send   │ │ /info       │ │ /view         │    │
│  │ /events │ │ /history│ │ /balance    │ │ /call         │    │
│  └────┬────┘ └────┬────┘ └──────┬──────┘ └───────┬───────┘    │
└───────┼──────────┼─────────────┼─────────────────┼───────────┘
        │          │             │                 │
        ▼          ▼             ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐   │
│  │ BlockService    │  │ TxSendService   │  │ ContractService│   │
│  │ TxService       │  │ TokenService    │  │ EventService  │   │
│  └────────┬────────┘  └────────┬────────┘  └───────┬───────┘   │
└───────────┼────────────────────┼───────────────────┼───────────┘
            │                    │                   │
            └────────────────────┼───────────────────┘
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                            │
│                    ┌─────────────────┐                        │
│                    │   EthClient     │                        │
│                    │ (go-ethereum)   │                        │
│                    └─────────────────┘                        │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    External Services                           │
│  ┌───────────────┐                ┌─────────────────────┐      │
│  │ Ethereum Node │                │      SQLite DB      │      │
│  │ (Sepolia/Local)│              │ (Tx History/Events) │      │
│  └───────────────┘                └─────────────────────┘      │
└─────────────────────────────────────────────────────────────────┘
```

### 包结构

```
go-ether/
├── api/                 # HTTP API 层
│   ├── handlers.go      # 基础处理器
│   ├── handlers_tx.go   # 交易相关处理器
│   ├── handlers_contract.go  # 合约相关处理器
│   └── http_server.go   # HTTP 服务器配置
├── client/              # Ethereum 客户端封装
│   ├── eth_client.go    # 基础客户端
│   └── multi_client.go  # 多网络客户端
├── config/              # 配置管理
│   ├── config.go        # 主配置
│   ├── network.go       # 网络配置
│   └── wallet.go        # 钱包配置
├── contract/            # 合约相关
│   ├── abi/             # ABI 文件
│   ├── bin/             # 字节码文件
│   └── bindings/        # abigen 生成的绑定代码
├── contracts/           # Solidity 合约源码
│   └── Counter.sol      # 计数器合约示例
├── pkg/                 # 公共工具包
│   └── converter/       # 单位转换工具
├── scripts/             # 脚本文件
│   └── gen_bindings.sh  # ABI 绑定生成脚本
├── service/             # 业务服务层
│   ├── block_service.go # 区块服务
│   ├── tx_service.go    # 交易查询服务
│   ├── tx_send_service.go # 交易发送服务
│   ├── event_service.go # 事件监听服务
│   ├── contract_service.go # 合约调用服务
│   └── token_service.go # 代币服务
├── store/               # 数据存储层
│   ├── memory_store.go  # 内存事件存储
│   └── tx_history_store.go # SQLite 交易历史存储
├── wallet/              # 钱包管理
│   └── signer.go        # 私钥签名器
├── main.go              # 主入口
└── go.mod               # Go 模块依赖
```

### 时序图

#### 区块查询流程

```
Client ──GET /api/block/{id}──▶ Handlers ──GetBlockByID──▶ BlockService ──BlockByNumber/Hash──▶ EthClient ──▶ Ethereum Node
   │                                      │                   │                          │
   │◀───JSON Response─────────────────────┼───────────────────┼──────────────────────────┘
   │                                      │                   │
   └───────────────────────────────────────┴───────────────────┘
```

#### 交易发送流程

```
Client ──POST /api/tx/send──▶ TxHandlers ──SendTransaction──▶ TxSendService
                                 │                                    │
                                 ├─▶ GetNonce ────────────────────────▶
                                 ├─▶ SuggestGasTipCap ────────────────▶
                                 ├─▶ SignTx ──────────────────────────▶
                                 ├─▶ SendTransaction ─────────────────▶ Ethereum Node
                                 └─▶ SaveToHistory ──────────────────▶ SQLite
```

## 🔌 API 接口说明

### 区块相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/block/{id}` | 查询区块 | `id`: 区块号或哈希 |

### 交易相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/tx/{hash}` | 查询交易 | `hash`: 交易哈希 |
| POST | `/api/tx/send` | 发送交易 | `{"to", "value", "gas", "data"}` |
| GET | `/api/tx/history` | 交易历史 | `limit`, `offset` |
| GET | `/api/tx/detail` | 交易详情 | `hash` |

### 事件相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/events` | 查询事件 | `limit`, `offset` |

### 合约相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| POST | `/api/contract/view` | 调用视图方法 | `{"contractAddr", "method", "args"}` |
| POST | `/api/contract/call` | 发送合约交易 | `{"contractAddr", "method", "args"}` |

### 代币相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/token/info` | 查询代币信息 | `token` |
| GET | `/api/token/balance` | 查询余额 | `token`, `holder` |
| POST | `/api/token/transfer` | 代币转账 | `{"tokenAddr", "to", "amount"}` |

## 🚀 本地安装与启动

### 环境要求

- Go 1.21+
- 以太坊节点（本地或远程）
- SQLite（已内置，无需额外安装）

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/dreamfsk/go-ether.git
cd go-ether

# 2. 安装依赖
go mod tidy

# 3. 编译
go build -o mini-block-explorer .
```

### 配置说明

创建 `.env` 文件或设置环境变量：

```env
# 网络配置
NETWORK=local                    # sepolia 或 local

# 以太坊节点配置
ETH_RPC_URL=http://localhost:8545
ETH_WS_URL=ws://localhost:8545

# ERC-20 合约地址（用于事件监听）
ERC20_CONTRACT=0xYourContractAddress

# 私钥（用于交易发送，可选）
SENDER_PRIVATE_KEY=0xYourPrivateKey
```

### 启动服务

```bash
# 方式1：使用 .env 文件
cp .env.example .env
# 编辑 .env 文件后启动
./mini-block-explorer

# 方式2：直接设置环境变量
export ETH_RPC_URL=http://localhost:8545
export ETH_WS_URL=ws://localhost:8545
export ERC20_CONTRACT=0x...
./mini-block-explorer
```

## 📝 使用示例

### 查询区块

```bash
curl http://localhost:8080/api/block/18500000
```

### 查询交易

```bash
curl http://localhost:8080/api/tx/0xabc123...
```

### 发送 ETH

```bash
curl -X POST http://localhost:8080/api/tx/send \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "value": "1000000000000000000"}'
```

### 查询代币信息

```bash
curl "http://localhost:8080/api/token/info?token=0xTokenAddress"
```

### 调用合约方法

```bash
# 视图方法（只读）
curl -X POST http://localhost:8080/api/contract/view \
  -H "Content-Type: application/json" \
  -d '{"contractAddr": "0xContractAddress", "method": "count"}'

# 写方法（需要签名）
curl -X POST http://localhost:8080/api/contract/call \
  -H "Content-Type: application/json" \
  -d '{"contractAddr": "0xContractAddress", "method": "increment"}'
```

## 🔧 配置部署说明

### 支持的网络

| 网络 | ChainID | 默认 RPC |
|------|---------|----------|
| Sepolia | 11155111 | `https://sepolia.infura.io/v3/YOUR_KEY` |
| Local | 31337 | `http://localhost:8545` |

### 切换网络

```bash
# 使用 Sepolia
export NETWORK=sepolia
export ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
export ETH_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY

# 使用本地链
export NETWORK=local
export ETH_RPC_URL=http://localhost:8545
export ETH_WS_URL=ws://localhost:8545
```

### 数据库配置

默认使用 SQLite 数据库，数据文件为 `transactions.db`。

### 代码规范

- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查代码
- 编写单元测试
- 遵循 Go 语言规范

## 📄 许可证

MIT License

## ❓ 常见问题

### Q: 连接以太坊节点失败？

**A**: 请确保：
1. 以太坊节点正在运行
2. RPC/WS URL 配置正确
3. 网络可达

### Q: 交易发送失败？

**A**: 请检查：
1. `SENDER_PRIVATE_KEY` 是否设置正确
2. 账户余额是否足够
3. Gas 价格是否合理

### Q: ERC-20 事件没有收到？

**A**: 请检查：
1. `ERC20_CONTRACT` 是否设置正确
2. 合约是否有 Transfer 事件
3. WebSocket 连接是否正常


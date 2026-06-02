# go-ether

一个基于 Go 语言和 go-ethereum 库构建的迷你区块浏览器和 ERC-20 监控服务。提供完整的区块链数据查询、交易发送、合约交互和代币管理功能。

使用 `abigen` 生成 MyERC20 合约的 Go 绑定代码，实现类型安全的合约交互。详细合约部署说明请参阅 [CONTRACT.md](./CONTRACT.md)。

## 功能特性

### 核心功能
- **区块查询**：支持通过区块号或哈希查询区块详情
- **交易查询**：查询交易详情、输入数据和回执信息
- **ERC-20 事件监听**：实时监听 Transfer 事件，持久化到 SQLite 数据库
- **交易发送**：支持 ETH 转账，自动处理 Gas 估算和交易签名
- **合约交互**：支持调用合约视图方法和发送合约交易
- **代币管理**：查询代币信息、余额，支持代币转账、铸造及部署
- **合约部署**：支持通过 API 部署 MyERC20 合约
- **历史追溯**：SQLite 统一存储，通过 `tx_type` 区分 ETH 转账和 ERC20 事件，支持按类型查询

### 架构特性
- **分层架构**：Client → Service → API / Store
- **多网络支持**：支持 Sepolia 测试网和本地测试链
- **类型安全绑定**：使用 abigen 生成的 Go 合约绑定，无需手动编码 ABI
- **统一持久化**：所有交易和事件数据统一存储到 SQLite，通过 `tx_type` 区分类型
- **优雅关闭**：支持 SIGINT/SIGTERM 信号处理
- **实时监控**：WebSocket 订阅 ERC-20 事件

## 架构设计

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP API Layer                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐ ┌───────────────┐    │
│  │ /api/   │ │ /api/tx │ │ /api/token  │ │ /api/contract │    │
│  │ block   │ │ /send   │ │ /info       │ │ /view         │    │
│  │ /events │ │ /history│ │ /balance    │ │ /call         │    │
│  │         │ │ /detail │ │ /transfer   │ │               │    │
│  │         │ │         │ │ /mint       │ │               │    │
│  │         │ │         │ │ /deploy     │ │               │    │
│  └────┬────┘ └────┬────┘ └──────┬──────┘ └───────┬───────┘    │
└───────┼──────────┼─────────────┼─────────────────┼───────────┘
        │          │             │                 │
        ▼          ▼             ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐   │
│  │ BlockService    │  │ TxSendService   │  │ ContractService│   │
│  │ TxService       │  │ ERC20Service    │  │ EventService  │   │
│  └────────┬────────┘  └────────┬────────┘  └───────┬───────┘   │
└───────────┼────────────────────┼───────────────────┼───────────┘
            │                    │                   │
            └────────────────────┼───────────────────┘
                                 ▼
                    ┌────────────────────────┐
                    │     SQLite 统一存储     │
                    │  (tx_history 表)       │
                    │  tx_type 区分:         │
                    │  - eth_transfer        │
                    │  - erc20_transfer      │
                    └────────────────────────┘
```

### 数据流

```
ETH 转账:  POST /api/tx/send → TxSendService → SQLite (tx_type=eth_transfer, status=pending)
                                              └→ 异步更新 status + block_number

ERC20 事件: WebSocket 监听 → EventService → SQLite (tx_type=erc20_transfer, status=success)

API 查询:
  GET /api/tx/history → List(limit, offset)       → 全部类型
  GET /api/events     → ListByType(erc20_transfer) → 仅 ERC20 事件
```

### 包结构

```
go-ether/
├── api/                     # HTTP API 层
│   ├── handlers.go          # 基础处理器（区块、交易、事件）
│   ├── handlers_tx.go       # ETH 交易发送处理器
│   ├── handlers_contract.go # 合约 & ERC20 处理器
│   └── http_server.go       # HTTP 服务器配置与路由注册
├── client/                  # Ethereum 客户端封装
│   └── eth_client.go        # 基础客户端
├── config/                  # 配置管理
│   ├── config.go            # 主配置（.env 加载）
│   └── network.go           # 网络参数配置
├── contracts/               # 合约源码与绑定
│   ├── MyERC20.sol          # ERC20 合约源码
│   └── myERC20.go           # abigen 生成的 Go 绑定代码
├── pkg/                     # 公共工具包
│   └── converter/           # 单位转换工具
├── service/                 # 业务服务层
│   ├── block_service.go     # 区块查询服务
│   ├── tx_service.go        # 交易查询服务
│   ├── tx_send_service.go   # ETH 交易发送服务（写入 SQLite）
│   ├── event_service.go     # ERC20 事件监听服务（写入 SQLite）
│   ├── contract_service.go  # 通用合约调用服务（动态 selector 缓存）
│   └── erc20_service.go     # ERC20 代币服务（基于 abigen 绑定，含部署）
├── store/                   # 数据存储层（SQLite 统一存储）
│   └── tx_history_store.go  # 交易/事件存储，支持按 tx_type 查询
├── tests/                   # 单元测试（统一测试目录）
├── wallet/                  # 钱包管理
│   └── signer.go            # 环境变量私钥签名器
├── compile.js               # solc 合约编译脚本
├── main.go                  # 主入口
├── go.mod                   # Go 模块依赖
├── package.json             # Node 依赖（solc + OpenZeppelin）
├── .env.example             # 环境变量配置模板
├── CONTRACT.md              # 合约部署和 abigen 详细指南
└── README.md                # 本文档
```

## API 接口说明

### 区块相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/block/{id}` | 查询区块 | `id`: 区块号或哈希 |

### 交易相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/tx/{hash}` | 查询链上交易 | `hash`: 交易哈希 |
| POST | `/api/tx/send` | 发送 ETH 交易 | `{"to", "value"}` |
| GET | `/api/tx/history` | 交易历史（含 ETH + ERC20） | `page`, `pageSize` |
| GET | `/api/tx/detail` | 本地交易详情 | `hash` |

### 事件相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/events` | 查询 ERC20 Transfer 事件 | -（返回最近 100 条 `erc20_transfer`） |

### 合约相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| POST | `/api/contract/view` | 调用视图方法（只读） | `{"contractAddr", "method", "args"}` |
| POST | `/api/contract/call` | 发送合约交易（写） | `{"contractAddr", "method", "args"}` |

### 代币相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/token/info` | 查询代币信息 | 无（使用 `.env` 配置的合约地址） |
| GET | `/api/token/balance` | 查询余额 | `holder` |
| POST | `/api/token/transfer` | 代币转账 | `{"to", "amount"}` |
| POST | `/api/token/mint` | 铸造代币 | `{"to", "amount"}` |
| POST | `/api/token/deploy` | 部署 MyERC20 合约 | `{"name", "symbol", "initialSupply", "recipient"}` |

### 数据字段说明

`tx_type` 取值：

| tx_type | 含义 | 写入来源 |
|---------|------|---------|
| `eth_transfer` | ETH 转账 | `TxSendService` 主动发送 |
| `erc20_transfer` | ERC20 代币 Transfer 事件 | `EventService` 链上监听 |

## 本地安装与启动

### 环境要求

- Go 1.21+
- Node.js（用于合约编译）
- 以太坊节点（本地或远程）
- SQLite（已内置，无需额外安装）

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/dreamfsk/go-ether.git
cd go-ether

# 2. 安装 Go 依赖
go mod tidy

# 3. 安装 abigen（Go 合约绑定工具）
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# 4. 安装 Node 依赖（solc + OpenZeppelin，用于编译合约）
npm install

# 5. 编译合约并生成 Go 绑定
node compile.js
abigen --abi=build/MyERC20.abi --bin=build/MyERC20.bin --pkg=contracts --out=contracts/myERC20.go

# 6. 编译项目
go build -o mini-block-explorer .
```

### 配置说明

创建 `.env` 文件（参考 `.env.example`）：

```env
# 网络配置
# 可选值: local, sepolia
NETWORK=sepolia

# 以太坊节点配置
ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
ETH_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY

# ERC-20 合约地址（已部署的 MyERC20 合约）
ERC20_CONTRACT=0xYourContractAddress

# 发送者私钥（用于签名交易，不要提交到版本控制）
SENDER_PRIVATE_KEY=your_private_key_here
```

### 启动服务

```bash
# 使用 .env 文件
./mini-block-explorer

# 或直接设置环境变量
NETWORK=sepolia ERC20_CONTRACT=0x... SENDER_PRIVATE_KEY=... ./mini-block-explorer
```

## 使用示例

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

### 查询交易历史

```bash
curl "http://localhost:8080/api/tx/history?page=1&pageSize=20"
```

### 查询 Transfer 事件

```bash
curl http://localhost:8080/api/events
```

### 查询代币信息

```bash
curl http://localhost:8080/api/token/info
```

### 查询代币余额

```bash
curl "http://localhost:8080/api/token/balance?holder=0xAddress"
```

### 代币转账

```bash
curl -X POST http://localhost:8080/api/token/transfer \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "amount": "1000000000000000000"}'
```

### 铸造代币

```bash
curl -X POST http://localhost:8080/api/token/mint \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "amount": "500000000000000000000"}'
```

### 部署合约

```bash
curl -X POST http://localhost:8080/api/token/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MyToken",
    "symbol": "MTK",
    "initialSupply": "1000000000000000000000",
    "recipient": "0xReceiverAddress"
  }'
```

返回示例：

```json
{
  "address": "0xDeployedContractAddress",
  "txHash": "0xTransactionHash"
}
```

### 调用合约方法

```bash
# 视图方法
curl -X POST http://localhost:8080/api/contract/view \
  -H "Content-Type: application/json" \
  -d '{"contractAddr": "0xContractAddress", "method": "name"}'

# 写方法（需要签名）
curl -X POST http://localhost:8080/api/contract/call \
  -H "Content-Type: application/json" \
  -d '{"contractAddr": "0xContractAddress", "method": "transfer", "args": ["0xTo", "1000000000000000000"]}'
```

## 网络配置

| 网络 | ChainID | 默认 RPC |
|------|---------|----------|
| Sepolia | 11155111 | `https://sepolia.infura.io/v3/YOUR_KEY` |
| Local | 31337 | `http://localhost:8545` |

### 切换网络

```bash
# Sepolia
export NETWORK=sepolia
export ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
export ETH_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY

# 本地链
export NETWORK=local
export ETH_RPC_URL=http://localhost:8545
export ETH_WS_URL=ws://localhost:8545
```

## 合约编译与绑定生成

详细说明请参阅 [CONTRACT.md](./CONTRACT.md)。

### 快速流程

```bash
# 1. 安装 Node 依赖
npm install

# 2. 编译 Solidity 合约
node compile.js

# 3. 生成 Go 绑定代码
abigen --abi=build/MyERC20.abi \
       --bin=build/MyERC20.bin \
       --pkg=contracts \
       --out=contracts/myERC20.go
```

### MyERC20 合约功能

| 方法 | 类型 | 描述 |
|------|------|------|
| `name()` | 只读 | 代币名称 |
| `symbol()` | 只读 | 代币符号 |
| `decimals()` | 只读 | 精度 |
| `totalSupply()` | 只读 | 总供应量 |
| `balanceOf(address)` | 只读 | 查询余额 |
| `allowance(address,address)` | 只读 | 查询授权额度 |
| `transfer(address,uint256)` | 写 | 转账 |
| `approve(address,uint256)` | 写 | 授权 |
| `transferFrom(address,address,uint256)` | 写 | 授权转账 |
| `mint(address,uint256)` | 写 | 铸造代币 |

## 运行测试

```bash
go test ./tests/...
```

## 常见问题

### Q: 连接以太坊节点失败？

A: 请确保：
1. 以太坊节点正在运行
2. RPC/WS URL 配置正确
3. 网络可达

### Q: 交易发送失败？

A: 请检查：
1. `SENDER_PRIVATE_KEY` 是否设置正确
2. 账户余额是否足够
3. Gas 价格是否合理

### Q: ERC-20 事件没有收到？

A: 请检查：
1. `ERC20_CONTRACT` 是否设置正确
2. 合约是否有 Transfer 事件
3. WebSocket 连接是否正常

### Q: abigen 命令未找到？

A: 
```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Q: solc 编译报错 File not found？

A:
```bash
npm install @openzeppelin/contracts
```

## 许可证

MIT License

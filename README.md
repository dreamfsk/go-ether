# go-ether

一个基于 Go 语言和 go-ethereum 库构建的迷你区块浏览器和 ERC-20 监控服务。提供完整的区块链数据查询、交易发送、合约交互和代币管理功能。

使用 `abigen` 生成 MyERC20 合约的 Go 绑定代码，实现类型安全的合约交互。详细合约部署说明请参阅 [CONTRACT.md](./CONTRACT.md)。

## 功能特性

### 核心功能
- **区块查询**：支持通过区块号或哈希查询区块详情
- **交易查询**：查询交易详情、输入数据和回执信息
- **ERC-20 事件监听**：实时监听 Transfer 事件，持久化到 SQLite 数据库，支持断线自动重连
- **交易发送**：支持 ETH 转账，自动处理 Gas 估算和交易签名
- **合约交互**：支持调用合约视图方法和发送合约交易
- **代币管理**：查询代币信息、余额，支持代币转账、铸造及部署
- **合约部署**：支持通过 API 部署 MyERC20 合约，部署后自动保存地址
- **合约地址管理**：支持多合约地址管理，可通过 API 切换当前合约，支持 SQLite 持久化
- **历史追溯**：SQLite 统一存储，通过 `tx_type` 区分 ETH 转账和 ERC20 事件，支持分页和地址过滤查询
- **前端管理面板**：React + TypeScript + Ant Design 构建的管理界面，支持查看合约列表、部署合约、切换监听合约等功能

### 架构特性
- **分层架构**：Client → Service → API / Store
- **多网络支持**：支持 Sepolia 测试网和本地测试链，EIP-155 签名适配（chainID 安全传递）
- **类型安全绑定**：使用 abigen 生成的 Go 合约绑定，ERC20 交互统一走 ERC20Service
- **Gas 动态估算**：所有 ERC20 写操作先调用 `eth_estimateGas`，失败时 1.5x 硬编码兜底
- **统一持久化**：所有交易和事件数据统一存储到 SQLite，通过 `tx_type` 区分类型
- **优雅关闭**：支持 SIGINT/SIGTERM 信号处理
- **实时监控**：WebSocket 订阅 ERC-20 事件，断线自动重连（指数退避）
- **RPC/WS 双通道**：HTTP RPC 用于合约调用/交易查询，WebSocket 专用事件订阅，WS 不可用时优雅降级
- **服务生命周期统一管理**：`ContractManager` 作为总控入口，所有 signer 相关 service 统一创建/销毁

## 架构设计

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP API Layer                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐ ┌───────────────┐    │
│  │ /api/   │ │ /api/tx │ │ /api/token  │ │ /api/contract │    │
│  │ block   │ │ /send   │ │ /info       │ │ /view         │    │
│  │ /events │ │ /history│ │ /balance    │ │ /call         │    │
│  │         │ │ /detail │ │ /transfer   │ │ /list         │    │
│  │         │ │         │ │ /mint       │ │ /switch       │    │
│  │         │ │         │ │ /deploy     │ │ /current      │    │
│  └────┬────┘ └────┬────┘ └──────┬──────┘ └───────┬───────┘    │
└───────┼──────────┼─────────────┼─────────────────┼───────────┘
        │          │             │                 │
        ▼          ▼             ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                           │
│  ┌─────┐ ┌────┐                        ┌──────────────────┐    │
│  │Block│ │ Tx │                        │ ContractManager  │    │
│  │Svc  │ │ Svc│                        │  ┌─────────────┐ │    │
│  └──┬──┘ └──┬─┘                        │  │TxSendService│ │    │
│     │       │                          │  │ERC20Service │ │    │
│     │       │                          │  │EventService │ │    │
│     │       │                          │  │ContractSvc  │ │    │
│     │       │                          │  └─────────────┘ │    │
│     │       │                          └────────┬─────────┘    │
│     │       │                                   │              │
│     │       │         ┌─────────────┐           │              │
│     │       └────────→│  SQLite     │←──────────┘              │
│     └────────────────→│  统一存储   │                          │
│                       │ tx_history │                          │
│                       │ contracts  │                          │
│                       └─────────────┘                          │
└─────────────────────────────────────────────────────────────────┘
```
**说明**: `ContractManager` 作为唯一的 signer 相关服务入口，统一管理 `TxSendService`、`ERC20Service`、`EventService`、`ContractService` 的生命周期。`/api/contract/view` 和 `/api/contract/call` 中的 ERC20 方法通过方法白名单自动委托到 `ERC20Service`（类型安全绑定），非 ERC20 方法预留扩展点。

### 数据流

```
ETH 转账:  POST /api/tx/send → TxSendService → SQLite (tx_type=eth_transfer, status=pending)
                                              └→ 异步更新 status + block_number

ERC20 事件: WebSocket 监听 → EventService → SQLite (tx_type=erc20_transfer, status=success)
                                                         └→ 断线自动重连（指数退避）

合约部署: POST /api/token/deploy → ERC20Service → ContractManager → SQLite (contracts 表)
                                                         └→ 自动设为活跃合约

合约切换: POST /api/contract/switch → ContractManager → 重建 ERC20Service/EventService/TxSendService

合约调用: POST /api/contract/view (或 call)
          └→ ERC20 方法? → handleERC20View/handleERC20Write → 动态创建 ERC20Service → 类型安全调用
          └→ 非 ERC20 方法 → 返回错误（预留扩展点）

API 查询:
  GET /api/tx/history → List(limit, offset)       → 全部类型
  GET /api/events     → ListByType(erc20_transfer) → 支持 address/limit/offset
```

### 包结构

```
go-ether/
├── api/                     # HTTP API 层
│   ├── handlers.go          # 基础处理器（区块、交易、事件）
│   ├── handlers_tx.go       # ETH 交易发送处理器
│   ├── handlers_contract.go # 合约 & ERC20 处理器（含地址管理）
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
│   ├── tx_send_service.go   # ETH 交易发送服务（含指数退避确认、写入 SQLite）
│   ├── event_service.go     # ERC20 事件监听服务（含自动重连）
│   ├── contract_service.go  # 通用合约调用骨架（selector 缓存，预留非 ERC20 扩展）
│   ├── erc20_service.go     # ERC20 代币服务（abigen 绑定、Gas 估算、类型安全调用、部署）
│   └── contract_manager.go  # 服务总控（统一管理 TxSend/ERC20/Event/Contract 生命周期）
├── store/                   # 数据存储层（SQLite 统一存储）
│   ├── tx_history_store.go  # 交易/事件存储，支持按 tx_type 查询
│   └── contract_store.go    # 合约地址存储
├── tests/                   # 单元测试（统一测试目录）
├── wallet/                  # 钱包管理
│   └── signer.go            # 环境变量私钥签名器
├── web/                     # 前端管理面板（React + TypeScript + Ant Design）
│   ├── src/
│   │   ├── api/             # API 客户端
│   │   ├── components/      # 公共组件
│   │   ├── layouts/         # 布局组件
│   │   ├── pages/           # 页面组件
│   │   └── types/           # TypeScript 类型定义
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
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
| GET | `/api/events` | 查询 ERC20 Transfer 事件 | `address`: 按地址过滤（可选），`limit`: 每页数量（默认 20），`offset`: 偏移量（默认 0） |

### 合约相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/contract/list` | 查询所有合约地址 | - |
| GET | `/api/contract/current` | 查询当前活跃合约 | - |
| POST | `/api/contract/switch` | 切换当前合约地址 | `{"address"}` |
| POST | `/api/contract/view` | 调用视图方法（ERC20 方法自动委托到 ERC20Service） | `{"contractAddr", "method", "args"}` |
| POST | `/api/contract/call` | 发送合约交易（ERC20 方法自动委托到 ERC20Service） | `{"contractAddr", "method", "args"}` |

**支持的 ERC20 方法**: `name`, `symbol`, `decimals`, `totalSupply`, `balanceOf`(address), `allowance`(owner,spender), `transfer`(to,amount), `mint`(to,amount), `approve`(spender,amount), `transferFrom`(from,to,amount)

### 代币相关

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | `/api/token/info` | 查询代币信息 | 无（使用当前活跃合约） |
| GET | `/api/token/balance` | 查询余额 | `holder` |
| POST | `/api/token/transfer` | 代币转账 | `{"to", "amount"}` |
| POST | `/api/token/mint` | 铸造代币 | `{"to", "amount"}` |
| POST | `/api/token/deploy` | 部署 MyERC20 合约（部署后自动设为活跃合约） | `{"name", "symbol", "initialSupply", "recipient"}` |

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

# 6. 构建前端管理面板
cd web
npm install
npm run build
cd ..

# 7. 编译项目
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

# ERC-20 合约地址（默认地址，仅在无保存地址时使用）
# 合约地址优先从 contracts.db 读取
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
# 分页查询
curl "http://localhost:8080/api/events?limit=20&offset=0"

# 按地址过滤
curl "http://localhost:8080/api/events?address=0x1234...&limit=20"
```

### 查询合约地址列表

```bash
curl http://localhost:8080/api/contract/list
```

### 查询当前活跃合约

```bash
curl http://localhost:8080/api/contract/current
```

### 切换合约地址

```bash
curl -X POST http://localhost:8080/api/contract/switch \
  -H "Content-Type: application/json" \
  -d '{"address": "0xNewContractAddress"}'
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
1. 当前活跃合约地址是否正确（`GET /api/contract/current`）
2. 合约是否有 Transfer 事件
3. WebSocket 连接是否正常（会自动重连）

### Q: 如何切换合约地址？

A: 使用 `POST /api/contract/switch` 接口：
```bash
curl -X POST http://localhost:8080/api/contract/switch \
  -H "Content-Type: application/json" \
  -d '{"address": "0xNewAddress"}'
```

### Q: 合约地址存储在哪里？

A: 合约地址持久化在 `contracts.db` SQLite 数据库中，启动时优先读取活跃地址。

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

## 前端管理面板

### 访问地址

启动服务后，访问前端管理面板：

```
http://localhost:8080/manage/index
```

### 页面功能

| 页面 | 路径 | 功能描述 |
|------|------|----------|
| 首页 | `/manage/index` | 显示统计概览、当前监听状态、部署合约入口 |
| 合约列表 | `/manage/contracts` | 查看所有部署的合约，支持切换监听、查看详情 |
| 合约详情 | `/manage/contracts/:address` | 查看合约详细信息、代币信息、事件记录 |

### 功能特性

1. **部署合约**：首页点击"部署合约"按钮，填写表单后一键部署，部署成功后自动加入监听
2. **合约管理**：查看所有合约列表，支持按名称、地址、网络筛选
3. **切换监听**：点击"切换监听"按钮，快速切换当前监听的合约
4. **状态监控**：实时显示当前监听状态、合约数量统计
5. **事件查看**：在合约详情页查看该合约的 Transfer 事件历史

## 许可证

MIT License

## 修复记录

详细问题分析及修复过程请参阅 [fix.md](./fix.md)。关键修复包括：

- **chainID 安全传递**：修复 3 处硬编码/遗漏的 EIP-155 签名错误
- **Gas 动态估算**：所有 ERC20 写操作增加 `eth_estimateGas` 估算
- **错误处理增强**：nonce 溢出防护、rows.Err() 检查、nil CallOpts 替换、错误日志统一
- **架构优化**：统一 ERC20 双路径到 ERC20Service、ContractManager 收拢服务生命周期、RPC/WS 双通道分离
- **接口封装**：Signer 接口新增 `TransactOpts`，消除私钥暴露
- **前端修复**：事件响应类型安全、合约详情页按地址查询 tokenInfo

# go-ether

一个基于 Go 语言和 go-ethereum 库构建的迷你区块浏览器和 ERC-20 监控服务。提供完整的区块链数据查询、交易发送、合约交互和代币管理功能。

使用 `abigen` 生成 MyERC20 合约的 Go 绑定代码，实现类型安全的合约交互。合约部署详情请参阅 [deploy/CONTRACT.md](./deploy/CONTRACT.md)。

## 功能特性

### 核心功能

- **区块查询**：支持通过区块号或哈希查询区块详情
- **交易查询**：查询交易详情、输入数据和回执信息
- **ERC-20 事件监听**：实时监听 Transfer 事件，持久化到 SQLite 数据库，支持断线自动重连
- **交易发送**：支持 ETH 转账和 ERC-20 代币转账，自动处理 Gas 估算和交易签名
- **合约交互**：支持调用合约视图方法和发送合约交易
- **代币管理**：查询代币信息、余额，支持代币转账、铸造及部署
- **合约部署**：支持通过 API 部署 MyERC20 合约，部署后自动保存地址
- **合约地址管理**：支持多合约地址管理，可通过 API 切换当前合约，支持 SQLite 持久化
- **历史追溯**：SQLite 统一存储，通过 `tx_type` 区分 ETH 转账和 ERC20 事件，支持分页和地址过滤查询
- **前端管理面板**：React + TypeScript + Ant Design 构建的管理界面，支持查看合约列表、部署合约、切换监听合约、发送交易等功能

### 架构特性

- **分层架构**：Client → Service → API / Store
- **多网络支持**：支持 Sepolia 测试网和本地测试链，EIP-155 签名适配（chainID 安全传递）
- **类型安全绑定**：使用 abigen 生成的 Go 合约绑定，ERC20 交互统一走 ERC20Service
- **Gas 动态估算**：所有 ERC20 写操作先调用 `eth_estimateGas`，失败时 1.5x 硬编码兜底
- **统一持久化**：所有交易和事件数据统一存储到 SQLite，通过 `tx_type` 区分类型
- **优雅关闭**：支持 SIGINT/SIGTERM 信号处理
- **实时监控**：WebSocket 订阅 ERC-20 事件，断线自动重连（指数退避）
- **RPC/WS 双通道**：HTTP RPC 用于合约调用/交易查询，WebSocket 专用事件订阅，WS 不可用时优雅降级
- **服务生命周期统一管理**：`ContractManager` 管理合约注册与切换，`ContractServiceBundle` 统一管理 `EventService`/`ERC20Service`/`TxSendService` 的生命周期

## 本地安装与启动

### 环境要求

- Go 1.21+
- Node.js 18+（推荐 20+）
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

# 4. 安装 Node 依赖（solc + OpenZeppelin，用于合约编译）
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

### 前端开发模式

如果需要修改前端代码：

```bash
cd web
npm install
npm run dev
```

前端开发服务器将在 `http://localhost:5173` 启动，支持热更新。

### 配置说明

创建 `.env` 文件（参考 `.env.example`）：

```env
# 网络配置
# 可选值: local, sepolia
NETWORK=local

# 以太坊节点配置
ETH_RPC_URL=http://localhost:8545
ETH_WS_URL=ws://localhost:8546

# ERC-20 合约地址（可选，启动后会自动从 contracts.db 读取）
# ERC20_CONTRACT=0xYourContractAddress

# 钱包配置（推荐 Keystore 方式）
# 方式一：使用 Keystore 文件（推荐，更安全）
# KEYSTORE_PATH=/home/user/.ethereum/keystore/UTC--2024-01-01--xxxxxx
# KEYSTORE_PASSWORD=your_keystore_password

# 方式二：使用环境变量私钥（仅供测试）
# SENDER_PRIVATE_KEY=your_private_key_here

# CORS 允许的前端域名白名单（逗号分隔，不设置时使用默认值）
# CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:8080

# 代币/ETH 操作金额上限（wei 单位，默认 10^30）
# MAX_TOKEN_AMOUNT=1000000000000000000000000000000
```

**注意**: Keystore 方式优先于环境变量方式。如果同时配置了 Keystore 和环境变量私钥，系统会优先使用 Keystore 方式。

### 启动服务

```bash
# 使用 .env 文件
./mini-block-explorer

# 或直接设置环境变量
NETWORK=sepolia ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_KEY KEYSTORE_PATH=/path/to/keystore KEYSTORE_PASSWORD=password ./mini-block-explorer
```

## 网络配置

| 网络 | ChainID | 默认 RPC | 默认 WS |
|------|---------|----------|---------|
| Sepolia | 11155111 | `https://sepolia.infura.io/v3/YOUR_KEY` | `wss://sepolia.infura.io/ws/v3/YOUR_KEY` |
| Local | 31337 | `http://127.0.0.1:8545` | `ws://127.0.0.1:8545` |

### 切换网络

```bash
# Sepolia
export NETWORK=sepolia
export ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
export ETH_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY

# 本地链
export NETWORK=local
export ETH_RPC_URL=http://127.0.0.1:8545
export ETH_WS_URL=ws://127.0.0.1:8545
```

## API 接口说明

参考文档：[docs/api.md](./docs/api.md)

### 区块相关

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/block/{id}` | 查询区块（区块号或哈希） |

### 交易相关

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/tx/{hash}` | 查询链上交易 |
| POST | `/api/tx/send` | 发送 ETH 交易 |
| GET | `/api/tx/history` | 交易历史（分页，`limit`/`offset`） |
| GET | `/api/tx/detail` | 本地交易详情（`hash` 参数） |
| GET | `/api/tx/gas-fee` | Gas 费用建议 |
| POST | `/api/tx/estimate-gas` | Gas 估算 |

### 事件相关

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/events` | 查询事件记录（`address`、`contractOnly`、`limit`、`offset`） |

> 当 `contractOnly=true` 时，仅查询 `contract_addr` 匹配的记录（ERC-20 事件）；不指定时查询所有 `from_addr`、`to_addr` 或 `contract_addr` 匹配的记录。

### 合约相关

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/contract/list` | 合约地址列表 |
| GET | `/api/contract/current` | 当前活跃合约 |
| POST | `/api/contract/switch` | 切换当前合约 |
| POST | `/api/contract/view` | 调用视图方法（ERC20 方法自动委托到 ERC20Service） |
| POST | `/api/contract/call` | 发送合约交易（ERC20 方法自动委托到 ERC20Service） |

**支持的 ERC20 方法**: `name`, `symbol`, `decimals`, `totalSupply`, `balanceOf`, `allowance`, `transfer`, `mint`, `approve`, `transferFrom`

### 代币相关

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/token/info` | 代币信息（`address` 可选） |
| GET | `/api/token/balance` | 查询余额（`holder` 必填，`address` 可选） |
| POST | `/api/token/transfer` | 代币转账，受限 `MAX_TOKEN_AMOUNT` |
| POST | `/api/token/mint` | 铸造代币，受限 `MAX_TOKEN_AMOUNT` |
| POST | `/api/token/deploy` | 部署新合约，部署后自动设为活跃 |

### 系统配置

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/config` | 查询系统配置（网络信息、Gas 估算、当前钱包地址等） |

### 数据字段说明

`tx_type` 取值：

| tx_type | 含义 | 写入来源 |
|---------|------|---------|
| `eth_transfer` | ETH 转账 | `TxSendService` 主动发送 |
| `erc20_transfer` | ERC20 代币 Transfer 事件 | `EventService` 链上监听 |

## 使用示例

### 区块与交易

```bash
# 查询区块
curl http://localhost:8080/api/block/18500000

# 查询链上交易
curl http://localhost:8080/api/tx/0xabc123...

# 查询交易历史
curl "http://localhost:8080/api/tx/history?limit=20&offset=0"

# 查询 Gas 费用建议
curl http://localhost:8080/api/tx/gas-fee

# 估算交易 Gas
curl -X POST http://localhost:8080/api/tx/estimate-gas \
  -H "Content-Type: application/json" \
  -d '{"from": "0xSender", "to": "0xReceiver", "value": "1000000000000000000"}'

# 查询事件记录
curl "http://localhost:8080/api/events?limit=20&offset=0"
curl "http://localhost:8080/api/events?address=0x1234...&limit=20"
```

### 发送 ETH

```bash
curl -X POST http://localhost:8080/api/tx/send \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "value": "1000000000000000000"}'
```

### 合约管理

```bash
# 查询合约列表
curl http://localhost:8080/api/contract/list

# 查询当前活跃合约
curl http://localhost:8080/api/contract/current

# 切换合约
curl -X POST http://localhost:8080/api/contract/switch \
  -H "Content-Type: application/json" \
  -d '{"address": "0xNewContractAddress"}'
```

### 代币操作

```bash
# 查询代币信息
curl http://localhost:8080/api/token/info
curl "http://localhost:8080/api/token/info?address=0xContractAddress"

# 查询余额
curl "http://localhost:8080/api/token/balance?holder=0xAddress"
curl "http://localhost:8080/api/token/balance?holder=0xAddress&address=0xContractAddress"

# 代币转账
curl -X POST http://localhost:8080/api/token/transfer \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "amount": "1000000000000000000"}'

# 铸造代币
curl -X POST http://localhost:8080/api/token/mint \
  -H "Content-Type: application/json" \
  -d '{"to": "0xReceiverAddress", "amount": "500000000000000000000"}'

# 部署合约
curl -X POST http://localhost:8080/api/token/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MyToken",
    "symbol": "MTK",
    "initialSupply": "1000000000000000000000",
    "recipient": "0xReceiverAddress"
  }'
# 返回示例: {"address": "0xDeployedContractAddress", "txHash": "0xTransactionHash"}
```

### 合约调用

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

## 合约编译与绑定生成

详细说明请参阅 [deploy/CONTRACT.md](./deploy/CONTRACT.md)。

### 快速流程

```bash
npm install
node compile.js
abigen --abi=build/MyERC20.abi --bin=build/MyERC20.bin --pkg=contracts --out=contracts/myERC20.go
```

### 一键部署脚本

```bash
cd deploy
./deploy_contract.sh
```

脚本自动完成合约编译、部署到链上、生成 Go 绑定代码。

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

## 前端管理面板

启动服务后，可通过以下方式访问：

- **生产构建**: `http://localhost:8080/manage/index`
- **开发模式**: `http://localhost:5173`（需先在 `web/` 目录执行 `npm run dev`）

### 页面功能

| 页面 | 路径 | 功能描述 |
|------|------|----------|
| 首页 | `/manage/index` | 统计概览、监听状态、部署合约入口 |
| 合约列表 | `/manage/contracts` | 查看所有合约，支持切换监听、查看详情 |
| 合约详情 | `/manage/contracts/:address` | 合约信息、代币信息、事件记录 |

### 功能特性

1. **部署合约**：首页点击"部署合约"按钮一键部署，部署成功后自动加入监听
2. **合约管理**：查看所有合约列表，支持按名称、地址、网络筛选
3. **切换监听**：点击"切换监听"按钮，快速切换当前监听的合约
4. **状态监控**：实时显示当前监听状态、合约数量统计
5. **事件查看**：在合约详情页查看该合约的 Transfer 事件历史
6. **发送交易**：支持 ETH 转账和 ERC-20 代币转账，可选择交易类型，自动填充合约地址，显示交易费用汇总

## 运行测试

```bash
go test ./tests/...
```

## 项目结构

```
go-ether/
├── api/                     # HTTP API 层
│   ├── handlers.go          # 基础处理器（区块、交易、事件）
│   ├── handlers_tx.go       # 交易发送、Gas 估算、历史查询
│   ├── handlers_contract.go # 合约 & ERC20 处理器
│   ├── http_server.go       # 路由注册与服务器配置
│   └── middleware.go        # CORS、日志、恢复中间件
├── build/                   # 合约编译输出（ABI / 字节码）
├── client/                  # Ethereum 客户端封装
│   ├── eth_client.go        # 基础客户端（HTTP/WebSocket）
│   └── multi_client.go      # 双通道客户端（RPC + WS）
├── config/                  # 配置管理
│   ├── config.go            # 主配置（.env 加载）
│   ├── network.go           # 网络参数（ChainID、RPC/WS URL）
│   └── wallet.go            # 钱包配置（Keystore / 私钥）
├── contracts/               # 合约源码与 Go 绑定
│   ├── MyERC20.sol          # ERC20 合约源码
│   └── myERC20.go           # abigen 生成的 Go 绑定
├── deploy/                  # 合约部署相关
│   ├── CONTRACT.md          # 合约部署指南
│   ├── deploy_contract.go   # 部署脚本（Go）
│   └── deploy_contract.sh   # 一键部署脚本
├── docs/                    # 项目文档
│   ├── ARCHITECTURE.md      # 架构设计文档
│   └── api.md               # API 详细文档
├── pkg/
│   └── converter/           # 单位转换工具（Wei / Ether）
├── service/                 # 业务服务层
│   ├── block_service.go     # 区块查询
│   ├── tx_service.go        # 交易查询、Gas 估算
│   ├── tx_send_service.go   # ETH 交易发送（指数退避确认、写入 SQLite）
│   ├── event_service.go     # ERC20 事件监听（WebSocket + 断线重连）
│   ├── erc20_service.go     # ERC20 代币服务（abigen 绑定、Gas 估算、部署）
│   ├── contract_service.go  # 通用合约调用（selector 缓存，预留扩展）
│   ├── contract_service_bundle.go # 服务生命周期管理
│   └── contract_manager.go  # 合约管理器（注册、切换、持久化）
├── store/                   # SQLite 存储层
│   ├── tx_history_store.go  # 交易/事件存储
│   └── contract_store.go    # 合约地址存储
├── tests/                   # 单元测试
├── wallet/                  # 钱包管理
│   ├── signer.go            # 环境变量私钥签名器
│   └── keystore.go          # Keystore 文件签名器
├── web/                     # 前端管理面板（React + TypeScript + Ant Design）
│   ├── public/
│   ├── src/
│   │   ├── api/             # API 客户端
│   │   ├── layouts/         # 布局组件
│   │   ├── pages/           # 页面（Home / ContractList / ContractDetail）
│   │   └── types/           # TypeScript 类型定义
│   ├── package.json
│   └── vite.config.ts
├── main.go                  # 主入口
├── compile.js               # solc 合约编译脚本
├── .env.example             # 环境变量模板
├── QUICKSTART.md            # 快速开始指南
└── README.md                # 本文档
```

## 架构设计

详情参考 [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)。

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP API Layer                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐ ┌───────────────┐    │
│  │ /api/   │ │ /api/tx │ │ /api/token  │ │ /api/contract │    │
│  │ block   │ │ /send   │ │ /info       │ │ /view         │    │
│  │ /events │ │ /history│ │ /balance    │ │ /call         │    │
│  │         │ │ /detail │ │ /transfer   │ │ /list         │    │
│  │         │ │ /gas-fee│ │ /mint       │ │ /switch       │    │
│  │         │ │ /estimate││ /deploy     │ │ /current      │    │
│  │         │ │ -gas    │ │             │ │               │    │
│  └────┬────┘ └────┬────┘ └──────┬──────┘ └───────┬───────┘    │
└───────┼──────────┼─────────────┼─────────────────┼───────────┘
        │          │             │                 │
        ▼          ▼             ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                           │
│  ┌─────┐ ┌────┐                        ┌──────────────────┐    │
│  │Block│ │ Tx │                        │ ContractManager  │    │
│  │Svc  │ │ Svc│                        │  ┌─────────────┐ │    │
│  └──┬──┘ └──┬─┘                        │  │ServiceBundle │ │    │
│     │       │                          │  │TxSendService │ │    │
│     │       │                          │  │ERC20Service  │ │    │
│     │       │                          │  │EventService  │ │    │
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

> `ContractManager` 负责合约注册与切换，`ContractServiceBundle` 统一管理 `TxSendService`、`ERC20Service`、`EventService` 的生命周期。ER20 方法通过方法白名单自动委托到 `ERC20Service`（类型安全绑定），非 ERC20 方法预留扩展点。

### 数据流

```
ETH 转账:  POST /api/tx/send → TxSendService → SQLite (tx_type=eth_transfer, status=pending)
                                              └→ 异步更新 status + block_number

ERC-20 代币转账: POST /api/token/transfer → ERC20Service → 合约调用 → SQLite (tx_type=erc20_transfer)
                                                                        └→ EventService 监听并记录 Transfer 事件

ERC20 事件: WebSocket 监听 → EventService → SQLite (tx_type=erc20_transfer, status=success)
                                                         └→ 断线自动重连（指数退避）

合约部署: POST /api/token/deploy → ERC20Service → ContractManager → SQLite (contracts 表)
                                                         └→ 自动设为活跃合约

合约切换: POST /api/contract/switch → ContractManager → ServiceBundle.Init() → 重建 EventService/ERC20Service/TxSendService

合约调用: POST /api/contract/view (或 call)
          └→ ERC20 方法? → handleERC20View/handleERC20Write → 动态创建 ERC20Service → 类型安全调用
          └→ 非 ERC20 方法 → 返回错误（预留扩展点）

API 查询:
  GET /api/tx/history → List(limit, offset)    → 全部类型
  GET /api/events     → ListByAddress()        → 按地址过滤，支持 contractOnly 参数
```

## 常见问题

**Q: 连接以太坊节点失败？**  
A: 请确保：以太坊节点正在运行，RPC/WS URL 配置正确，网络可达。

**Q: 交易发送失败？**  
A: 请检查：钱包私钥/Keystore 设置正确，账户余额充足，Gas 价格合理。

**Q: ERC-20 事件没有收到？**  
A: 请检查：当前活跃合约地址是否正确（`GET /api/contract/current`），合约是否有 Transfer 事件，WebSocket 连接是否正常（会自动重连）。

**Q: 如何切换合约地址？**  
A: 使用 `POST /api/contract/switch` 接口：
```bash
curl -X POST http://localhost:8080/api/contract/switch \
  -H "Content-Type: application/json" \
  -d '{"address": "0xNewAddress"}'
```

**Q: 合约地址存储在哪里？**  
A: 合约地址持久化在 `contracts.db` SQLite 数据库中，启动时优先读取活跃地址。

**Q: abigen 命令未找到？**  
A:
```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

**Q: solc 编译报错 File not found？**  
A:
```bash
npm install @openzeppelin/contracts
```

## 待扩展优化

- [ ] `transactions.db` 和 `contracts.db` 在 `main.go` 中硬编码，改为从环境变量或配置文件读取或换成数据库如 MySQL、PostgreSQL 等
- [ ] 添加 JWT 认证

## 许可证

MIT License

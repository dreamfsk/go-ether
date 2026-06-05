# 架构设计文档

## 1. 系统概述

go-ether 是一个基于 Go 语言和 go-ethereum 库构建的迷你区块浏览器和 ERC-20 监控服务。系统采用分层架构，实现区块数据查询、交易发送、合约交互和代币管理的完整功能。

### 1.1 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Go 1.21+ |
| 以太坊交互 | go-ethereum |
| 前端框架 | React 19 + TypeScript 6 |
| UI 组件库 | Ant Design 6 |
| 构建工具 | Vite 8 |
| 合约编译 | solc 0.8.20 |
| 数据存储 | SQLite |

### 1.2 网络支持

| 网络 | ChainID | RPC |
|------|---------|-----|
| Sepolia | 11155111 | Infura/Alchemy |
| Local | 31337 | localhost:8545 |

---

## 2. 架构设计

### 2.1 分层架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP API Layer (api/)                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │   Block     │ │     Tx      │ │   Token     │ │ Contract  │ │
│  │  Handler    │ │  Handler    │ │  Handler    │ │  Handler  │ │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────┬─────┘ │
└─────────┼───────────────┼───────────────┼──────────────┼───────┘
          │               │               │              │
          ▼               ▼               ▼              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Service Layer (service/)                    │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐│
│  │   Block     │ │     Tx      │ │      ContractManager       ││
│  │  Service    │ │  Service    │ │  ┌───────────────────────┐ ││
│  └─────────────┘ └─────────────┘ │  │  ContractServiceBundle │ ││
│                                │  │  ┌─────────────────┐   │ ││
│                                │  │  │  ERC20Service   │   │ ││
│                                │  │  │  EventService   │   │ ││
│                                │  │  │  TxSendService  │   │ ││
│                                │  │  └─────────────────┘   │ ││
│                                │  └───────────────────────┘ ││
│                                └─────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
          │               │               │
          ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Client Layer (client/)                      │
│  ┌─────────────────────┐      ┌─────────────────────────────┐ │
│  │      EthClient      │      │        MultiClient          │ │
│  │  HTTP / WebSocket   │      │  RPC (HTTP) + WS 双通道     │ │
│  └─────────────────────┘      └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Ethereum Network                          │
│                    Sepolia / Local Network                      │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 包结构

```
go-ether/
├── api/                          # HTTP API 层
│   ├── handlers.go              # 基础处理器（区块、交易、事件）
│   ├── handlers_tx.go           # ETH 交易发送处理器
│   ├── handlers_contract.go    # 合约 & ERC20 处理器
│   └── http_server.go          # HTTP 服务器配置与路由
│
├── client/                       # Ethereum 客户端封装
│   ├── eth_client.go           # 基础客户端（HTTP/WebSocket）
│   └── multi_client.go         # 双通道客户端（RPC + WS）
│
├── config/                       # 配置管理
│   ├── config.go               # 主配置（.env 加载）
│   └── network.go             # 网络参数配置
│
├── contracts/                    # 合约源码与绑定
│   ├── MyERC20.sol            # ERC20 合约源码
│   └── myERC20.go             # abigen 生成的 Go 绑定
│
├── service/                     # 业务服务层
│   ├── block_service.go       # 区块查询服务
│   ├── tx_service.go          # 交易查询服务
│   ├── tx_send_service.go     # ETH 交易发送服务
│   ├── event_service.go       # ERC20 事件监听服务
│   ├── erc20_service.go       # ERC20 代币服务
│   ├── contract_service.go    # 通用合约调用
│   ├── contract_service_bundle.go  # 服务生命周期管理
│   └── contract_manager.go    # 合约管理器
│
├── store/                       # 数据存储层
│   ├── tx_history_store.go    # 交易/事件存储
│   └── contract_store.go      # 合约地址存储
│
├── wallet/                      # 钱包管理
│   ├── signer.go              # 签名器接口
│   └── keystore.go           # Keystore 钱包实现
│
├── pkg/                         # 公共工具包
│   └── converter/             # 单位转换
│
└── web/                         # 前端管理面板
    └── src/
        ├── api/               # API 客户端
        ├── pages/            # 页面组件
        └── types/            # 类型定义
```

---

## 3. 核心组件设计

### 3.1 MultiClient（双通道客户端）

MultiClient 封装 HTTP RPC 和 WebSocket 连接，提供统一接口。

```go
type MultiClient struct {
    rpc *EthClient   // HTTP RPC 客户端
    ws  *EthClient   // WebSocket 客户端
}
```

**职责**：
- RPC 通道：用于合约调用、交易查询、余额查询等只读操作
- WS 通道：专用事件订阅（Transfer 事件监听）
- 优雅降级：WS 不可用时仅禁用事件监听，不影响其他功能

### 3.2 ContractManager（合约管理器）

ContractManager 是服务层的核心组件，负责：
- 合约注册与切换
- 委托 ServiceBundle 管理服务生命周期
- 持久化合约地址到 SQLite

```go
type ContractManager struct {
    multiClient   *client.MultiClient
    signer        wallet.Signer
    contractStore *store.ContractStore
    txHistory     *store.TxHistoryStore
    serviceBundle *ContractServiceBundle
}
```

### 3.3 ContractServiceBundle（服务生命周期管理）

ServiceBundle 统一管理三个服务的生命周期：

```go
type ContractServiceBundle struct {
    erc20Service   *ERC20Service   // ERC20 代币交互
    eventService   *EventService   // 事件监听
    txSendService  *TxSendService  // ETH 转账
}
```

**生命周期管理**：
- `Init(address)`: 初始化/切换合约，重建所有服务
- `StartListening()`: 启动事件监听
- `Stop()`: 停止所有服务

### 3.4 ERC20Service（ERC20 代币服务）

基于 abigen 生成的 Go 绑定，提供类型安全的合约交互：

```go
type ERC20Service struct {
    client       *client.EthClient
    signer       wallet.Signer
    contract     *contracts.MyERC20   // abigen 绑定
    contractAddr common.Address
    chainID      *big.Int
}
```

**核心方法**：
- `GetTokenInfo()`: 查询代币基本信息
- `BalanceOf()`: 查询余额
- `Transfer()`: 转账
- `Mint()`: 铸造
- `Deploy()`: 部署合约

### 3.5 EventService（事件监听服务）

通过 WebSocket 订阅 ERC-20 Transfer 事件：

```go
type EventService struct {
    client       *client.EthClient
    contractAddr common.Address
    fromBlock    *big.Int
    txHistory    *store.TxHistoryStore
}
```

**特性**：
- 实时监听 Transfer 事件
- 断线自动重连（指数退避）
- 事件持久化到 SQLite

---

## 4. 数据流设计

### 4.1 ETH 转账

```
POST /api/tx/send
    │
    ▼
TxSendService.SendETH()
    │
    ├── 签名交易
    ├── 发送交易
    └── 写入 SQLite (tx_type=eth_transfer, status=pending)
              │
              ▼
         异步等待确认
              │
              ▼
         更新 status=success, block_number
```

### 4.2 ERC-20 代币转账

```
POST /api/token/transfer
    │
    ▼
ERC20Service.Transfer()
    │
    ├── 签名交易
    ├── 发送交易
    ├── 写入 SQLite (tx_type=erc20_transfer)
    │
    ▼
EventService 监听 Transfer 事件
    │
    ▼
更新事件记录
```

### 4.3 合约部署

```
POST /api/token/deploy
    │
    ▼
ERC20Service.Deploy()
    │
    ├── 部署合约
    ├── ContractManager.AddDeployedContract()
    │
    ├── 写入 contracts.db
    ├── 设置为活跃合约
    │
    ▼
ServiceBundle.Init() → 重建所有服务
    │
    ▼
自动开始监听新合约
```

### 4.4 合约切换

```
POST /api/contract/switch
    │
    ▼
ContractManager.SwitchContract()
    │
    ├── 更新 contracts.db (IsActive)
    │
    ▼
ServiceBundle.Init() → 停止旧服务，创建新服务
    │
    ▼
StartListening() → 开始监听新合约事件
```

---

## 5. 存储设计

### 5.1 SQLite 表结构

**transactions 表** - 统一存储交易和事件

```sql
CREATE TABLE transactions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    hash        TEXT NOT NULL UNIQUE,
    from_addr   TEXT,
    to_addr     TEXT,
    value       TEXT,
    gas_used    INTEGER,
    status      TEXT DEFAULT 'pending',
    block_number INTEGER,
    tx_type     TEXT NOT NULL,      -- eth_transfer / erc20_transfer
    contract_addr TEXT,             -- ERC20 合约地址
    timestamp   DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**contracts 表** - 合约地址管理

```sql
CREATE TABLE contracts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    address     TEXT NOT NULL UNIQUE,
    name        TEXT,
    symbol      TEXT,
    network     TEXT,
    deployer    TEXT,
    tx_hash     TEXT,
    is_active   INTEGER DEFAULT 0,
    created_at  DATETIME
);
```

### 5.2 tx_type 字段说明

| tx_type | 来源 | 说明 |
|---------|------|------|
| `eth_transfer` | TxSendService | 主动发送的 ETH 转账 |
| `erc20_transfer` | EventService | 链上监听的 Transfer 事件 |

---

## 6. API 路由设计

### 6.1 路由分组

```
/api/
├── block/                    # 区块查询
│   └── {id}                 # GET - 查询区块
│
├── tx/                      # 交易相关
│   ├── {hash}              # GET - 查询交易
│   ├── send                # POST - 发送 ETH
│   ├── history             # GET - 交易历史
│   └── detail              # GET - 本地交易详情
│
├── events                   # GET - 查询事件
│
├── token/                   # 代币相关
│   ├── info                # GET - 代币信息
│   ├── balance             # GET - 余额查询
│   ├── transfer            # POST - 代币转账
│   ├── mint                # POST - 铸造代币
│   └── deploy              # POST - 部署合约
│
└── contract/               # 合约相关
    ├── list                # GET - 合约列表
    ├── current             # GET - 当前合约
    ├── switch              # POST - 切换合约
    ├── view                # POST - 视图调用
    └── call                # POST - 写调用
```

### 6.2 合约方法白名单

系统通过白名单机制区分 ERC20 方法和非 ERC20 方法：

```go
var erc20Methods = map[string]bool{
    "name":         true,
    "symbol":       true,
    "decimals":     true,
    "totalSupply":  true,
    "balanceOf":    true,
    "allowance":    true,
    "transfer":     true,
    "approve":      true,
    "transferFrom": true,
    "mint":         true,
}
```

**处理逻辑**：
- ERC20 方法 → 自动委托到 ERC20Service（类型安全）
- 非 ERC20 方法 → 返回错误（预留扩展点）

---

## 7. 配置管理

### 7.1 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `NETWORK` | 网络类型 | `sepolia` / `local` |
| `ETH_RPC_URL` | HTTP RPC 地址 | `https://sepolia.infura.io/v3/KEY` |
| `ETH_WS_URL` | WebSocket 地址 | `wss://sepolia.infura.io/ws/v3/KEY` |
| `ERC20_CONTRACT` | 默认合约地址 | `0x...` |
| `SENDER_PRIVATE_KEY` | 私钥 | `0x...` |
| `KEYSTORE_PATH` | Keystore 路径 | `/path/to/keystore` |
| `KEYSTORE_PASSWORD` | Keystore 密码 | `password` |
| `MAX_TOKEN_AMOUNT` | 操作金额上限 | `1000000000000000000000000000000` |

### 7.2 钱包配置优先级

```
Keystore 方式 > 环境变量私钥方式
```

---

## 8. 安全性设计

### 8.1 钱包安全

- 支持 Keystore 文件加载（推荐）
- 支持环境变量私钥（仅用于开发/测试）
- 私钥不持久化存储

### 8.2 金额限制

- 所有代币/ETH 操作受 `MAX_TOKEN_AMOUNT` 限制
- 防止意外的大额转账

### 8.3 路径安全

- 前端静态文件服务防止路径穿越攻击
- 合约地址严格校验格式

---

## 9. 错误处理

### 9.1 错误响应格式

```json
{
    "error": "错误描述",
    "code": "ERROR_CODE"
}
```

### 9.2 错误码

| 错误码 | 说明 |
|--------|------|
| `INVALID_ADDRESS` | 无效的以太坊地址 |
| `INSUFFICIENT_BALANCE` | 余额不足 |
| `CONTRACT_ERROR` | 合约调用失败 |
| `SIGNER_NOT_FOUND` | 未配置签名钱包 |
| `NETWORK_ERROR` | 网络连接错误 |

---

## 10. 扩展点

### 10.1 待实现功能

- [ ] transactions.db 和 contracts.db 路径从环境变量读取或换成其它数据库（如MySQL）
- [ ] JWT 认证

### 10.2 预留扩展点

- 非 ERC20 合约方法支持
- 多代币支持
- 自定义事件类型监听

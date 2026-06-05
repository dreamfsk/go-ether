# go-ether API 文档

## 概述

go-ether 是一个基于 Go 语言和 go-ethereum 库构建的区块链浏览器和 ERC-20 代币管理服务。本文档描述了所有可用的 API 接口。

**基础 URL**: `http://localhost:8080/api`

---

## 目录

- [区块相关](#区块相关)
- [交易相关](#交易相关)
- [合约管理](#合约管理)
- [ERC20 代币](#erc20-代币)
- [系统配置](#系统配置)

---

## 区块相关

### 获取区块信息

获取指定区块的详细信息。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/block/{id}` |
| 参数 | `id` - 区块号或区块哈希 |

**请求示例**:

```bash
curl http://localhost:8080/api/block/1234567
```

**响应示例** (200 OK):

```json
{
  "number": 1234567,
  "hash": "0x...",
  "parentHash": "0x...",
  "transactions": [...],
  "gasUsed": "12345678",
  "timestamp": 1234567890
}
```

**错误响应** (404 Not Found):

```json
{
  "error": "block not found"
}
```

---

## 交易相关

### 获取交易详情

通过交易哈希获取交易详情。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/tx/{hash}` |
| 参数 | `hash` - 交易哈希 |

**请求示例**:

```bash
curl http://localhost:8080/api/tx/0x123456...
```

**响应示例** (200 OK):

```json
{
  "hash": "0x123456...",
  "from": "0x...",
  "to": "0x...",
  "value": "1000000000000000000",
  "gasLimit": 21000,
  "gasPrice": "1000000000",
  "status": "confirmed"
}
```

### 获取事件列表

获取历史交易事件列表，支持按地址筛选。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/events` |
| 查询参数 | `address` (可选) - 地址筛选<br>`contractOnly` (可选) - 仅合约交互<br>`limit` (可选, 默认 20) - 分页数量<br>`offset` (可选, 默认 0) - 偏移量 |

**请求示例**:

```bash
curl http://localhost:8080/api/events?address=0x...&limit=10&offset=0
```

**响应示例** (200 OK):

```json
{
  "events": [
    {
      "hash": "0x...",
      "from": "0x...",
      "to": "0x...",
      "value": "...",
      "type": "transfer",
      "timestamp": 1234567890
    }
  ],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### 发送 ETH 交易

发送 ETH 转账交易。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/tx/send` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "to": "0x123456...",
  "value": "1000000000000000000",
  "gasLimit": 21000,
  "gasPrice": "1000000000"
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/tx/send \
  -H "Content-Type: application/json" \
  -d '{"to":"0x...","value":"1000000000000000000"}'
```

**响应示例** (200 OK):

```json
{
  "txHash": "0x123456...",
  "status": "pending"
}
```

### 获取交易历史

获取本地记录的交易历史。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/tx/history` |
| 查询参数 | `limit` (可选, 默认 10) - 数量<br>`offset` (可选, 默认 0) - 偏移量 |

**请求示例**:

```bash
curl http://localhost:8080/api/tx/history?limit=20&offset=0
```

**响应示例** (200 OK):

```json
[
  {
    "hash": "0x...",
    "from": "0x...",
    "to": "0x...",
    "value": "...",
    "status": "confirmed",
    "timestamp": 1234567890
  }
]
```

### 获取交易详情

通过哈希获取本地记录的交易详情。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/tx/detail` |
| 查询参数 | `hash` - 交易哈希 |

**请求示例**:

```bash
curl http://localhost:8080/api/tx/detail?hash=0x123456...
```

**响应示例** (200 OK):

```json
{
  "hash": "0x...",
  "from": "0x...",
  "to": "0x...",
  "value": "...",
  "status": "confirmed",
  "timestamp": 1234567890
}
```

### 获取 Gas 费用建议

获取当前网络的 Gas 费用建议。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/tx/gas-fee` |

**请求示例**:

```bash
curl http://localhost:8080/api/tx/gas-fee
```

**响应示例** (200 OK):

```json
{
  "low": "1000000000",
  "medium": "2000000000",
  "high": "3000000000"
}
```

### 估算 Gas

估算执行交易所需的 Gas 数量。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/tx/estimate-gas` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "from": "0x...",
  "to": "0x...",
  "value": "1000000000000000000"
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/tx/estimate-gas \
  -H "Content-Type: application/json" \
  -d '{"to":"0x...","value":"1000000000000000000"}'
```

**响应示例** (200 OK):

```json
{
  "gas": 21000
}
```

---

## 合约管理

### 列出已部署合约

获取所有已部署的合约列表。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/contract/list` |

**请求示例**:

```bash
curl http://localhost:8080/api/contract/list
```

**响应示例** (200 OK):

```json
{
  "contracts": [
    {
      "address": "0x...",
      "name": "MyToken",
      "symbol": "MTK"
    }
  ],
  "currentAddress": "0x..."
}
```

### 切换当前合约

切换当前活动合约。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/contract/switch` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "address": "0x123456..."
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/contract/switch \
  -H "Content-Type: application/json" \
  -d '{"address":"0x..."}'
```

**响应示例** (200 OK):

```json
{
  "success": true,
  "address": "0x...",
  "message": "Contract switched successfully"
}
```

### 获取当前合约

获取当前活动合约的信息。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/contract/current` |

**请求示例**:

```bash
curl http://localhost:8080/api/contract/current
```

**响应示例** (200 OK):

```json
{
  "address": "0x...",
  "name": "MyToken",
  "symbol": "MTK"
}
```

### 合约视图调用

调用合约的只读方法。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/contract/view` |
| Content-Type | `application/json` |

**支持的方法**: `name`, `symbol`, `decimals`, `totalSupply`, `balanceOf`, `allowance`

**请求体**:

```json
{
  "method": "balanceOf",
  "args": ["0x..."],
  "contractAddr": "0x..."
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/contract/view \
  -H "Content-Type: application/json" \
  -d '{"method":"balanceOf","args":["0x..."],"contractAddr":"0x..."}'
```

**响应示例** (200 OK):

```json
{
  "status": "success",
  "result": "1000000000000000000"
}
```

### 合约写调用

调用合约的写方法（需要签名）。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/contract/call` |
| Content-Type | `application/json` |

**支持的方法**: `transfer`, `mint`, `approve`, `transferFrom`

**请求体**:

```json
{
  "method": "transfer",
  "args": ["0x...", "1000000000000000000"],
  "contractAddr": "0x..."
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/contract/call \
  -H "Content-Type: application/json" \
  -d '{"method":"transfer","args":["0x...","1000000000000000000"],"contractAddr":"0x..."}'
```

**响应示例** (200 OK):

```json
{
  "status": "pending",
  "txHash": "0x123456..."
}
```

---

## ERC20 代币

### 获取代币信息

获取当前或指定合约的代币信息。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/token/info` |
| 查询参数 | `address` (可选) - 指定合约地址 |

**请求示例**:

```bash
curl http://localhost:8080/api/token/info?address=0x...
```

**响应示例** (200 OK):

```json
{
  "name": "MyToken",
  "symbol": "MTK",
  "decimals": 18,
  "totalSupply": "1000000000000000000000",
  "contractAddr": "0x..."
}
```

### 查询代币余额

查询指定地址的代币余额。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/token/balance` |
| 查询参数 | `holder` - 持有地址<br>`address` (可选) - 指定合约地址 |

**请求示例**:

```bash
curl http://localhost:8080/api/token/balance?holder=0x...
```

**响应示例** (200 OK):

```json
{
  "holder": "0x...",
  "balance": "1000000000000000000",
  "contractAddr": "0x..."
}
```

### 代币转账

发送代币转账交易。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/token/transfer` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "to": "0x...",
  "amount": "1000000000000000000"
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/token/transfer \
  -H "Content-Type: application/json" \
  -d '{"to":"0x...","amount":"1000000000000000000"}'
```

**响应示例** (200 OK):

```json
{
  "txHash": "0x123456...",
  "status": "pending"
}
```

### 铸造代币

铸造新代币（需要权限）。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/token/mint` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "to": "0x...",
  "amount": "1000000000000000000"
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/token/mint \
  -H "Content-Type: application/json" \
  -d '{"to":"0x...","amount":"1000000000000000000"}'
```

**响应示例** (200 OK):

```json
{
  "txHash": "0x123456...",
  "status": "pending"
}
```

### 部署新合约

部署新的 ERC20 代币合约。

| 属性 | 值 |
|------|-----|
| 方法 | POST |
| 路径 | `/token/deploy` |
| Content-Type | `application/json` |

**请求体**:

```json
{
  "name": "MyToken",
  "symbol": "MTK",
  "initialSupply": "1000000000000000000000",
  "recipient": "0x..."
}
```

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/token/deploy \
  -H "Content-Type: application/json" \
  -d '{"name":"MyToken","symbol":"MTK","initialSupply":"1000000000000000000000"}'
```

**响应示例** (200 OK):

```json
{
  "address": "0x123456...",
  "txHash": "0xabc123...",
  "status": "pending"
}
```

---

## 系统配置

### 获取系统配置

获取系统配置信息。

| 属性 | 值 |
|------|-----|
| 方法 | GET |
| 路径 | `/config` |

**请求示例**:

```bash
curl http://localhost:8080/api/config
```

**响应示例** (200 OK):

```json
{
  "maxTokenAmount": "1000000000000000000000000000000"
}
```

---

## 错误响应

所有 API 接口在出错时返回相应的 HTTP 状态码和错误信息。

### 常见错误码

| 状态码 | 描述 |
|--------|------|
| 400 | 无效的请求参数 |
| 404 | 资源未找到 |
| 405 | 方法不允许 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

**错误响应格式**:

```json
{
  "error": "详细错误描述"
}
```

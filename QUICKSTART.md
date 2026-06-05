# 快速开始

## 前置准备

确保已安装：
- Go 1.21+
- Node.js 18+（推荐 20+）

---

## 后端配置与启动

### 1. 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装 abigen 工具
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# 安装 Node 依赖（用于合约编译）
npm install
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入：

```bash
NETWORK=sepolia
ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
ETH_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY
SENDER_PRIVATE_KEY=your_private_key_here
```

### 3. 编译合约（可选）

如果需要部署新合约：

```bash
node compile.js
abigen --abi=build/MyERC20.abi --bin=build/MyERC20.bin --pkg=contracts --out=contracts/myERC20.go
```

### 4. 启动后端服务

```bash
go run main.go
```

后端服务将在 `http://localhost:8080` 启动。

---

## 前端配置与启动

### 1. 安装前端依赖

```bash
cd web
npm install
```

### 2. 启动前端开发服务器

```bash
npm run dev
```

前端将在 `http://localhost:5173` 启动。

### 3. 构建生产版本（可选）

```bash
npm run build
```

构建后的文件将放置在 `web/dist` 目录中，可通过后端静态文件服务访问。

---

## 访问应用

启动后端和前端后：

- **前端管理面板**: `http://localhost:5173`（开发模式）或 `http://localhost:8080/manage/index`（生产构建）
- **API 文档**: `http://localhost:8080/api`（详见 `docs/api.md`）
- **合约部署**: 详见 `deploy/CONTRACT.md`

---

## 快速验证

### 1. 检查后端健康

```bash
curl http://localhost:8080/api/block/latest
```

### 2. 访问前端

打开浏览器访问 `http://localhost:5173` 或 `http://localhost:8080/manage/index`

---

## 常见问题

### abigen 命令未找到

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### 前端连接后端失败

确保后端已在 `http://localhost:8080` 启动，并且 CORS 配置正确。

### 合约部署失败

检查：
1. RPC URL 配置正确
2. 私钥有效且账户有余额
3. 使用 Sepolia 测试网络

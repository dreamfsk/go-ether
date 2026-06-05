# MyERC20 合约部署指南

编译合约 → 部署合约 → abigen生成Go绑定

## 快速开始 - 一键部署

### 1. 环境准备

确保已安装以下依赖：

```bash
# Node.js 依赖
npm install

# 安装 abigen 工具
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并填入你的配置：

```bash
cd ..
cp .env.example .env
```

编辑 `.env` 文件，填入：

```bash
# Infura RPC URL
ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY

# 部署者私钥
SENDER_PRIVATE_KEY=your_private_key_here
```

### 3. 运行一键部署脚本

```bash
cd deploy
./deploy_contract.sh
```

---

## 部署流程说明

`deploy_contract.sh` 会自动完成以下步骤：

1. **编译合约** - 编译 `contracts/MyERC20.sol` 生成 ABI 和字节码
2. **部署合约** - 部署合约到 Sepolia 测试网
3. **生成 Go 绑定** - 使用 abigen 生成 Go 合约绑定代码

---

## 部署后

脚本执行完成后，你会得到：

- 合约地址（在控制台输出）
- 交易哈希（可在 Etherscan 查看）
- Go 绑定代码（`contracts/myERC20.go`）

将合约地址添加到 `.env` 文件）：
可选，用于后续服务启动后作为默认合约地址，快速体验合约功能。

```bash
ERC20_CONTRACT=0xYourContractAddress
```

然后启动服务：

```bash
go run main.go
```

---

## 手动部署（可选）

如果你需要单独执行某个步骤：

```bash
# 1. 编译合约
node compile.js

# 2. 部署合约
cd deploy
go run deploy_contract.go

# 3. 生成 Go 绑定
abigen --abi=build/MyERC20.abi \
       --bin=build/MyERC20.bin \
       --pkg=contracts \
       --out=contracts/myERC20.go
```

---

## Sepolia 测试网资源

- **水龙头**: https://sepoliafaucet.com/
- **区块浏览器**: https://sepolia.etherscan.io/
- **Alchemy 水龙头**: https://www.alchemy.com/faucets/ethereum-sepolia

---

## 故障排查

### abigen 命令未找到

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### 编译报错 File not found

确保已安装依赖：

```bash
npm install
```

### 部署失败 "insufficient funds"

从 Sepolia 水龙头获取测试 ETH。

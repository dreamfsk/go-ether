# MyERC20 合约部署与绑定指南

## 概述

本项目包含一个基于 OpenZeppelin 的 ERC20 代币合约 `MyERC20.sol`，支持标准 ERC20 功能及 `mint` 铸造接口。本文档涵盖合约编译、Sepolia 测试网部署及 `abigen` Go 绑定代码生成的完整流程。

## 目录结构

```
go-ether/
├── contracts/
│   ├── MyERC20.sol          # 合约源码
│   └── myERC20.go           # abigen 生成的 Go 绑定代码
├── build/                    # solc 编译产物（已加入 .gitignore）
│   ├── MyERC20.abi
│   └── MyERC20.bin
├── compile.js                # solc 编译脚本
├── package.json              # Node 依赖（solc、OpenZeppelin）
└── CONTRACT.md               # 本文档
```

---

## 一、环境准备

### 1.1 安装 Node.js 依赖（solc + OpenZeppelin）

```bash
npm install
```

这将安装：
- `solc`（v0.8.20+）—— Solidity 编译器
- `@openzeppelin/contracts`（v5.x）—— OpenZeppelin 合约库

### 1.2 安装 abigen（Go 合约绑定生成工具）

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

验证安装：

```bash
abigen --version
```

### 1.3 准备 Sepolia 账户

- 一个持有 Sepolia ETH 的以太坊账户（可从 [Sepolia Faucet](https://sepoliafaucet.com/) 获取测试币）
- 账户私钥
- Infura / Alchemy 等 RPC Provider 的 API Key

---

## 二、编译合约（solc）

使用 `compile.js` 脚本编译 `MyERC20.sol`：

```bash
node compile.js
```

脚本执行流程：
1. 读取 `contracts/MyERC20.sol` 源码
2. 通过自定义 import 解析器加载 `@openzeppelin/contracts` 依赖
3. 使用 solc 编译（开启 optimizer，runs=200）
4. 输出 ABI 到 `build/MyERC20.abi`
5. 输出字节码到 `build/MyERC20.bin`

成功输出示例：

```
Contract data keys: [ 'MyERC20' ]
✓ ABI saved to build/MyERC20.abi
✓ Bytecode saved to build/MyERC20.bin

✓ Compilation completed successfully!

Contract ABI functions:
  - name()
  - symbol()
  - decimals()
  - totalSupply()
  - balanceOf(address)
  - transfer(address,uint256)
  - allowance(address,address)
  - approve(address,uint256)
  - transferFrom(address,address,uint256)
  - mint(address,uint256)
```

---

## 三、部署合约到 Sepolia 测试网

### 3.1 配置环境变量

创建 `.env` 文件（参考 `.env.example`）：

```bash
# 网络配置
NETWORK=sepolia

# Infura RPC URL（使用你的 Infura API Key 替换 YOUR_INFURA_KEY）
ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY

# ERC20 合约地址（部署后填入）
ERC20_CONTRACT=0xDeployedContractAddress

# 部署者私钥（不要提交到版本控制）
SENDER_PRIVATE_KEY=your_private_key_here
```

### 3.2 方式一：使用 Remix IDE 部署（推荐新手）

1. 打开 [Remix IDE](https://remix.ethereum.org/)
2. 在 `File Explorer` 中创建 `MyERC20.sol`，粘贴合约代码
3. 进入 `Solidity Compiler` 面板：
   - 选择编译器版本 `0.8.20+commit.a1b79de6`
   - 勾选 `Enable optimization`，runs 设为 `200`
   - 点击 `Compile MyERC20.sol`
4. 进入 `Deploy & Run Transactions` 面板：
   - 环境选择 `Injected Provider - MetaMask`
   - MetaMask 切换到 **Sepolia 测试网**
   - 合约选择 `MyERC20`
   - 展开 `Deploy`，填入构造函数参数：
     - `NAME`: `MyToken`
     - `SYMBOL`: `MTK`
     - `INITIAL_SUPPLY`: `1000000000000000000000`（1000 代币，18 位精度）
     - `RECIPIENT`: `0xYourAddress`
   - 点击 `Transact`，MetaMask 确认交易
5. 复制已部署合约的地址

### 3.3 方式二：使用 Go 代码部署

部署逻辑已集成在 `contracts/myERC20.go` 中的 `DeployMyERC20` 函数。可通过项目 API 或独立脚本调用：

```go
package main

import (
    "context"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/meu/go-ether/contracts"
)

func main() {
    // 连接 Sepolia
    client, _ := ethclient.Dial("https://sepolia.infura.io/v3/YOUR_KEY")
    
    // 创建交易选项（使用部署者私钥）
    privateKey, _ := crypto.HexToECDSA("your_private_key")
    chainID := big.NewInt(11155111) // Sepolia chain ID
    auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)

    // 部署合约
    addr, tx, contract, err := contracts.DeployMyERC20(
        auth,
        client,
        "MyToken",                          // 代币名称
        "MTK",                             // 代币符号
        big.NewInt(1000_000000000000000000), // 初始供应量（1000 代币）
        auth.From,                          // 接收者
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("合约地址: %s", addr.Hex())
    log.Printf("交易哈希: %s", tx.Hash().Hex())
    _ = contract
}
```

### 3.4 方式三：使用 Hardhat/Foundry 脚本部署

#### Hardhat

```javascript
// scripts/deploy.js
const hre = require("hardhat");

async function main() {
    const MyERC20 = await hre.ethers.getContractFactory("MyERC20");
    const contract = await MyERC20.deploy(
        "MyToken",
        "MTK",
        hre.ethers.parseEther("1000"),
        "0xRecipientAddress"
    );
    await contract.waitForDeployment();
    console.log("部署完成:", await contract.getAddress());
}

main().catch(console.error);
```

```bash
npx hardhat run scripts/deploy.js --network sepolia
```

#### Foundry

```bash
forge create --rpc-url $SEPOLIA_RPC_URL \
    --private-key $PRIVATE_KEY \
    contracts/MyERC20.sol:MyERC20 \
    --constructor-args "MyToken" "MTK" 1000000000000000000000 0xYourAddress
```

---

## 四、生成 Go 绑定代码（abigen）

编译合约并生成 ABI/Bytecode 后，使用 `abigen` 生成 Go 绑定代码：

### 4.1 完整命令

```bash
abigen --abi=build/MyERC20.abi \
       --bin=build/MyERC20.bin \
       --pkg=contracts \
       --out=contracts/myERC20.go
```

### 4.2 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `--abi` | 合约 ABI 文件路径 | `build/MyERC20.abi` |
| `--bin` | 合约部署字节码文件路径 | `build/MyERC20.bin` |
| `--pkg` | 生成 Go 代码的包名 | `contracts` |
| `--out` | 输出文件路径 | `contracts/myERC20.go` |
| `--type` | 合约类型名（可选，默认取合约名） | `MyERC20` |

### 4.3 生成的代码结构

`abigen` 会为合约生成以下类型和方法：

```
MyERC20                     # 完整合约绑定（组合 Caller + Transactor + Filterer）
MyERC20Caller               # 只读调用（name, symbol, balanceOf 等）
MyERC20Transactor           # 写交易（transfer, approve, mint 等）
MyERC20Filterer             # 事件过滤（Transfer, Approval 事件）
MyERC20Session              # Session 绑定（预配置交易选项）
DeployMyERC20()             # 合约部署函数
MyERC20MetaData             # ABI 和字节码元数据
```

### 4.4 获取合约实例（连接到已部署合约）

```go
import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/meu/go-ether/contracts"
)

// contractAddr 为部署后的合约地址
contractAddr := "0xDeployedContractAddress"
erc20, err := contracts.NewMyERC20(common.HexToAddress(contractAddr), ethClient)
```

### 4.5 一键脚本

编译 + 生成绑定的完整流程：

```bash
# 1. 编译合约
node compile.js

# 2. 生成 Go 绑定
abigen --abi=build/MyERC20.abi \
       --bin=build/MyERC20.bin \
       --pkg=contracts \
       --out=contracts/myERC20.go
```

---

## 五、Sepolia 网络配置参考

### 5.1 Sepolia 网络参数

| 参数 | 值 |
|------|-----|
| Chain ID | `11155111` |
| RPC URL | `https://sepolia.infura.io/v3/{YOUR_KEY}` |
| 区块浏览器 | https://sepolia.etherscan.io |
| 水龙头 | https://sepoliafaucet.com/ |

### 5.2 项目中的网络配置

项目已在 `config/network.go` 中预置 Sepolia 配置：

```go
NetworkSepolia: {
    Name:    "sepolia",
    RPCURL:  "https://sepolia.infura.io/v3/YOUR_INFURA_KEY",
    WSURL:   "wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY",
    ChainID: big.NewInt(11155111),
}
```

### 5.3 .env 配置示例

```bash
NETWORK=sepolia
ETH_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
ERC20_CONTRACT=0x你的合约地址
SENDER_PRIVATE_KEY=你的私钥（不含0x前缀）
```

---

## 六、验证合约部署

### 6.1 通过项目 API 验证

启动服务后调用 API：

```bash
# 查询代币信息
curl http://localhost:8080/api/token/info
```

返回示例：

```json
{
  "name": "MyToken",
  "symbol": "MTK",
  "decimals": 18,
  "totalSupply": "1000000000000000000000",
  "contractAddr": "0x..."
}
```

```bash
# 查询余额
curl "http://localhost:8080/api/token/balance?holder=0x你的地址"
```

### 6.2 通过 Etherscan 验证

1. 访问 https://sepolia.etherscan.io/
2. 搜索你的合约地址
3. 在 `Contract` 标签页验证并发布合约源码
4. 验证后可在线调用合约方法

---

## 七、故障排查

### solc 编译报错 File not found

确保已安装 OpenZeppelin 依赖：

```bash
npm install @openzeppelin/contracts
```

### abigen 命令未找到

```bash
# 安装 abigen
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
```

### MetaMask 部署失败 "insufficient funds"

从 Sepolia 水龙头获取测试币：
- https://sepoliafaucet.com/
- https://www.alchemy.com/faucets/ethereum-sepolia

### Go 编译报错 type mismatch

重新生成绑定代码确保与当前合约一致：

```bash
node compile.js && \
  abigen --abi=build/MyERC20.abi --bin=build/MyERC20.bin --pkg=contracts --out=contracts/myERC20.go
```

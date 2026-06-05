#!/bin/bash

set -e

echo "========================================"
echo "  MyERC20 合约一键部署流程"
echo "========================================"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# 1. 编译合约
echo "[1/3] 编译合约..."
node compile.js
echo ""

# 2. 部署合约
echo "[2/3] 部署合约到 Sepolia..."
cd deploy
go run deploy_contract.go
cd ..
echo ""

# 3. 生成 Go 绑定代码
echo "[3/3] 生成 Go 绑定代码..."
abigen --abi=build/MyERC20.abi \
       --bin=build/MyERC20.bin \
       --pkg=contracts \
       --out=contracts/myERC20.go
echo "✓ Go 绑定代码已生成: contracts/myERC20.go"
echo ""

# 完成
echo "========================================"
echo "  部署完成！"
echo "========================================"
echo ""
echo "下一步："
echo "  1. 将部署后的合约地址添加到 .env 文件"
echo "  2. 运行 go run main.go 启动服务"
echo ""

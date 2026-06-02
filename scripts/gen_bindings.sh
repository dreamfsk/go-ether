#!/bin/bash

set -e

echo "=== 生成合约绑定代码 ==="

CONTRACT_NAME="Counter"
CONTRACTS_DIR="contracts"
BUILD_DIR="build"
ABI_DIR="contract/abi"
BIN_DIR="contract/bin"
BINDINGS_DIR="contract/bindings"

mkdir -p $BUILD_DIR
mkdir -p $ABI_DIR
mkdir -p $BIN_DIR
mkdir -p $BINDINGS_DIR

echo "1. 编译合约..."
solc --abi "$CONTRACTS_DIR/$CONTRACT_NAME.sol" -o $BUILD_DIR --overwrite
solc --bin "$CONTRACTS_DIR/$CONTRACT_NAME.sol" -o $BUILD_DIR --overwrite

echo "2. 复制 ABI 和字节码..."
cp "$BUILD_DIR/$CONTRACT_NAME.abi" "$ABI_DIR/$CONTRACT_NAME.json"
cp "$BUILD_DIR/$CONTRACT_NAME.bin" "$BIN_DIR/$CONTRACT_NAME.bin"

echo "3. 生成 Go 绑定代码..."
abigen --abi="$BUILD_DIR/$CONTRACT_NAME.abi" \
       --bin="$BUILD_DIR/$CONTRACT_NAME.bin" \
       --pkg=bindings \
       --out="$BINDINGS_DIR/$CONTRACT_NAME.go"

echo "4. 清理临时文件..."
rm -rf $BUILD_DIR

echo "=== 绑定代码生成完成 ==="
echo "ABI: $ABI_DIR/$CONTRACT_NAME.json"
echo "字节码: $BIN_DIR/$CONTRACT_NAME.bin"
echo "绑定代码: $BINDINGS_DIR/$CONTRACT_NAME.go"
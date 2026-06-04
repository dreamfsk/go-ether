const solc = require('solc');
const fs = require('fs');
const path = require('path');

const contractPath = path.join(__dirname, 'contracts', 'MyERC20.sol');
const outputDir = path.join(__dirname, 'build');
const nodeModules = path.join(__dirname, 'node_modules');

// 创建输出目录
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir);
}

// 读取合约源码
const source = fs.readFileSync(contractPath, 'utf8');

// 创建 solc 输入
const input = {
    language: 'Solidity',
    sources: {
        'MyERC20.sol': {
            content: source
        }
    },
    settings: {
        optimizer: {
            enabled: true,
            runs: 200
        },
        outputSelection: {
            '*': {
                '*': ['abi', 'evm.bytecode']
            }
        }
    }
};

// 自定义 import 解析器
function findImport(importPath) {
    // 处理 OpenZeppelin 合约
    if (importPath.startsWith('@openzeppelin/')) {
        const filePath = path.join(nodeModules, importPath);
        if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf8');
            return { contents: content };
        }
    }
    return { error: 'File not found: ' + importPath };
}

// 使用本地 solc
const solcSnapshot = solc;
const compiled = solcSnapshot.compile(JSON.stringify(input), {
    import: findImport
});

const output = JSON.parse(compiled);

// 处理编译错误
if (output.errors) {
    const errors = output.errors.filter(e => e.severity === 'error');
    if (errors.length > 0) {
        errors.forEach(error => {
            console.error(error.formattedMessage);
        });
        process.exit(1);
    }
}

// 查找 MyERC20 合约
const contracts = output.contracts || {};
const myERC20Data = contracts['MyERC20.sol'];

if (!myERC20Data) {
    console.error('Compilation failed: MyERC20 contract not found');
    console.log('Available contracts:', Object.keys(contracts));
    process.exit(1);
}

// 检查合约数据的结构
console.log('Contract data keys:', Object.keys(myERC20Data));

// 新版本 solc 可能返回 { MyERC20: { abi, evm } } 结构
const contractInfo = myERC20Data.MyERC20 || Object.values(myERC20Data)[0];

if (!contractInfo) {
    console.error('Cannot find contract info in:', myERC20Data);
    process.exit(1);
}

const contractABI = contractInfo.abi;
const bytecode = contractInfo.evm?.bytecode?.object;

// 保存 ABI
fs.writeFileSync(
    path.join(outputDir, 'MyERC20.abi'),
    JSON.stringify(contractABI, null, 2)
);
console.log('✓ ABI saved to build/MyERC20.abi');

// 保存字节码
if (bytecode) {
    fs.writeFileSync(
        path.join(outputDir, 'MyERC20.bin'),
        bytecode
    );
    console.log('✓ Bytecode saved to build/MyERC20.bin');
} else {
    console.log('⚠ No bytecode found (might be an interface/abstract contract)');
}

console.log('\n✓ Compilation completed successfully!');
console.log('\nContract ABI functions:');
contractABI.forEach(item => {
    if (item.type === 'function') {
        console.log(`  - ${item.name}(${item.inputs.map(i => i.type).join(', ')})`);
    }
});

# go代理

go env

go env -w GOPROXY=https://goproxy.cn,direct ; 

go env -w GOSUMDB=sum.golang.google.cn

配置详情：

- GOPROXY : https://goproxy.cn,direct - 使用七牛云国内代理，如果代理不可用则直接回源

- GOSUMDB : sum.golang.google.cn - 使用国内校验服务器
  说明：
- goproxy.cn 是一个国内镜像，提供稳定的 Go 模块拉取服务

- ,direct 表示当代理无法访问时会直接从源站下载

- sum.golang.google.cn 是 sum.golang.org 的国内镜像


wsl -d Ubuntu * 操作文件系统
          
由于 WSL 命令执行存在环境问题，如果操作失败，先把文件保存到 Windows 公共目录C:\Users\Public\<project_name_tem>，之后再复制到 WSL 项目中

参考代码\\wsl$\Ubuntu\home\meu\projects\hardhat3nft\contracts\nft\MetaNFT.sol 你生成的\\wsl$\Ubuntu\home\meu\projects\hardhat3nft\contracts\nft\BasicNFT.sol。比较一下代码。然后是你生成的有问题。Invalid contract specified in override list: "ERC721".(2353)。你可能需要用wsl -d Ubuntu * 相关命令操作文件系统，只需要告诉我怎么改。由于 WSL 命令执行存在环境问题，如果操作失败，先把文件保存到 Windows 公共目录C:\Users\Public\<project_name_tem>，之后再复制到 WSL 项目中

分析一下测试日志问题：你可能需要用wsl -d Ubuntu * 相关命令。只需要告诉我怎么改

npx hardhat test test/AuctionTransparent.test.ts --grep "should upgrade contract successfully" --verbose

# 项目创建

## 
```bash
go mod init github.com/meu/go-ether"
```
### 3.1 安装 `go-ethereum`

在示例工程中引入模块依赖：

```bash
go get github.com/ethereum/go-ethereum
```

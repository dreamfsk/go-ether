查询 Sepolia 测试网交易信息最常用的方式有两种：

---

### 方式一：区块浏览器（最简单）

直接把交易哈希拼到 URL 后面：

```
https://sepolia.etherscan.io/tx/0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4f4105483eb5c
```

你刚才那笔交易打开就能看到：发送方、接收方、金额、Gas 费、区块确认数等全部信息。

> 地址也可以用同样的方式查看余额和交易历史：
> ```
> https://sepolia.etherscan.io/address/0xcfC4A9866E3B28F6DE3159E1413D481174bd5fB5
> ```

---

### 方式二：用 Foundry `cast` 命令行查询

你在 `.env` 里已经配了 Sepolia RPC，可以直接用 cast 查：

```bash
# 查询交易详情
cast tx 0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4f4105483eb5c \
  --rpc-url https://ethereum-sepolia.publicnode.com

# 查询交易收据
cast receipt 0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4f4105483eb5c \
  --rpc-url https://ethereum-sepolia.publicnode.com

# 查询地址余额
cast balance 0xcfC4A9866E3B28F6DE3159E1413D481174bd5fB5 \
  --rpc-url https://ethereum-sepolia.publicnode.com
```

---

### 你的项目中已有 `--tx` 查询功能

你刚才已经在用的 `go run txops.go --tx <txHash>` 就是自己项目里的交易查询功能，走的是 RPC 直查，和 cast 底层一样，只是输出格式更精简。
# test01
Starting: I:\go_path\bin\dlv.exe dap --listen=127.0.0.1:59264 from I:\solidy_space\go-ether
DAP server listening at: 127.0.0.1:59264
Building I:\solidy_space\go-ether\main.go
Type 'dlv help' for list of commands.
=== Ethereum Node Info ===
RPC URL       : https://ethereum-sepolia.publicnode.com
Chain ID      : 11155111

⚠️  注意: 'Latest' 区块是节点当前认为的最新区块，可能尚未被所有节点确认
   不同RPC节点可能返回不同的 'latest' 区块，导致与浏览器不匹配
   建议对比 'Safe' 或 'Finalized' 区块（已确认的区块）

Latest Block  : 10950682
Block Hash    : 0xb1814d7061ce9bc5d5a4c00a9c1000b69469d1959a58fad9f16150c972474c74
Block Time    : 2026-05-30T08:23:36+08:00
==========================
Prev Block    : 10950681 (0x71dd00a844334e9c57c3b49c18e9dcfd9ed0dbf74cb27fcc786de4e3e93b4f5d)

=== Safe Block (推荐对比) ===
Block Number  : 10950637
Block Hash    : 0x6a9eaf46927702424c5a116cba6ea1f5a97ca1539c57f4db955ba195c5ff8712 (RPC提供的hash，与浏览器一致)
Calculated    : 0xd9160afcce8df251ed4df974722bc90b07d3081481c762ee6f89756a638b123d (计算出的hash，可能不匹配)
Block Time    : 2026-05-30T08:14:24+08:00
Confirmations : 45
=============================

=== Finalized Block ===
Block Number  : 10950608
Block Hash    : 0xf1664e626b5e28bfede2fd27a78da6db81686d8595add2bac17fc796960302f2 (RPC提供的hash，与浏览器一致)
Calculated    : 0xae23c31c84952d271a487dbc6f8f3a9eb2e952259efadf98502815ae538eea7a (计算出的hash，可能不匹配)
Block Time    : 2026-05-30T08:08:00+08:00
Confirmations : 74
========================
Process 20456 has exited with status 0
Detaching
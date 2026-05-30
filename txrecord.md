I:\solidy_space\go-ether\03-tx-ops>go run txops.go --send --to 0x42cba1ac37d868781465ba03a71f --amount 0.1
=== Transaction Sent ===
From       : 0xcfC4A9866E3B28F6DE3159E1413D481174bd5fB5
To         : 0x6bd860745f756336eB3Aa92868781465BA03A71f
Value      : 0.100000 ETH (100000000000000000 Wei)
Gas Limit  : 21000
Gas Tip Cap: 1440000 Wei
Gas Fee Cap: 2634116176 Wei
Nonce      : 25
Tx Hash    : 0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4f4105483e

Transaction is pending. Use --tx flag to query status:
  go run main.go --tx 0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4

I:\solidy_space\go-ether\03-tx-ops>go run txops.go --tx 0xa30752011cd9c92ffb105483eb5c
=== Transaction ===
Hash        : 0xa30752011cd9c92ffb3ccde05c0c9717f858b35cb5a1dde8ddd4f4105483eb5c
Nonce       : 25
Gas         : 21000
Gas Price   : 2634116176
To          : 0x6bd860745f756336eB3Aa92868781465BA03A71f
Value (Wei) : 100000000000000000
Data Len    : 0 bytes
Pending     : false
=== Receipt ===
Status      : 1
BlockNumber : 10951002
BlockHash   : 0x619b5b9115e3e4d5eb357db7dde7ab17f5fd1b35c127883f73241c52d42d27c2
TxIndex     : 195
Gas Used    : 21000
Logs        : 0
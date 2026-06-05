package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	client, _ := ethclient.Dial(os.Getenv("ETH_RPC_URL"))
	chainID, _ := client.NetworkID(context.Background())
	privateKey, _ := crypto.HexToECDSA(os.Getenv("SENDER_PRIVATE_KEY"))
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)

	bytecode, _ := os.ReadFile("../build/MyERC20.bin")
	abiJSON, _ := os.ReadFile("../build/MyERC20.abi")
	parsedABI, _ := abi.JSON(strings.NewReader(string(abiJSON)))

	initialSupply, _ := new(big.Int).SetString("1000000000000000000000", 10)

	address, tx, _, _ := bind.DeployContract(
		auth, parsedABI, common.FromHex(string(bytecode)), client,
		"MyToken", "MTK", initialSupply, crypto.PubkeyToAddress(privateKey.PublicKey),
	)

	log.Printf("合约地址: %s", address.Hex())
	log.Printf("交易: https://sepolia.etherscan.io/tx/%s", tx.Hash().Hex())
}

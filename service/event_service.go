package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/store"
)

const erc20ABIJSON = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

type EventService struct {
	client   *client.EthClient
	store    *store.EventStore
	contract common.Address
	abi      abi.ABI
}

func NewEventService(c *client.EthClient, s *store.EventStore, contractAddr string) (*EventService, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &EventService{
		client:   c,
		store:    s,
		contract: common.HexToAddress(contractAddr),
		abi:      parsedABI,
	}, nil
}

func (s *EventService) StartListening(ctx context.Context) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{s.contract},
	}

	logsCh := make(chan types.Log)
	sub, err := s.client.SubscribeFilterLogs(ctx, query, logsCh)
	if err != nil {
		log.Printf("failed to subscribe logs: %v", err)
		return
	}
	defer sub.Unsubscribe()

	log.Printf("listening Transfer events of %s", s.contract.Hex())

	for {
		select {
		case vLog := <-logsCh:
			s.processLog(vLog)
		case err := <-sub.Err():
			log.Printf("subscription error: %v", err)
			return
		case <-ctx.Done():
			log.Println("context cancelled, stop subscription")
			return
		}
	}
}

func (s *EventService) processLog(vLog types.Log) {
	if len(vLog.Topics) == 0 {
		return
	}

	var event struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	if err := s.abi.UnpackIntoInterface(&event, "Transfer", vLog.Data); err != nil {
		log.Printf("failed to unpack log data: %v", err)
		return
	}

	if len(vLog.Topics) >= 3 {
		event.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		event.To = common.BytesToAddress(vLog.Topics[2].Bytes())
	}

	s.store.Add(store.TransferEvent{
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash.Hex(),
		From:        event.From.Hex(),
		To:          event.To.Hex(),
		Value:       event.Value.String(),
		Timestamp:   time.Now(),
	})
}

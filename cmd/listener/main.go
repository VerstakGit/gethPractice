package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gethPractice/contracts"
)

type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

func main() {
	endpoint := flag.String("endpoint", "", "endpoint of eth node")
	flag.Parse()
	c, err := ethclient.Dial(*endpoint)
	if err != nil {
		log.Println("eth client dial err", err)
		return
	}
	defer c.Close()

	address := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	logs := make(chan types.Log)
	sub, err := c.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{address},
	}, logs)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(contracts.TokenMetaData.ABI))
	if err != nil {
		log.Fatal(err)
	}

	logTransferSig := []byte("Transfer(address,address,uint256)")
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case logEntry := <-logs:
			if bytes.Equal(logEntry.Topics[0].Bytes(), logTransferSig) {
				continue
			}
			var transferEvent LogTransfer
			err = contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", logEntry.Data)
			if err != nil {
				log.Println("unpack err:", err)
				continue
			}
			transferEvent.From = common.HexToAddress(logEntry.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(logEntry.Topics[2].Hex())
			log.Printf("from: %v, to: %v, amount: %v", transferEvent.From.Hex(), transferEvent.To.Hex(), transferEvent.Tokens.String())
		}
	}
}

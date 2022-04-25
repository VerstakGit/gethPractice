package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	endpoint := flag.String("endpoint", "", "endpoint of an eth node")
	targetAddress := flag.String("address", "0x7f268357A8c2552623316e2562D90e642bB538E5", "contract address of an exchange")
	flag.Parse()

	rpcClient, err := rpc.Dial(*endpoint)
	if err != nil {
		log.Fatalf("rpc dial err: %v", err)
	}
	defer rpcClient.Close()

	subClient := gethclient.New(rpcClient)
	client, err := ethclient.Dial(*endpoint)
	if err != nil {
		log.Fatalf("eth client dial err: %v", err)
	}
	defer client.Close()

	exchangeContract := common.HexToAddress(*targetAddress).Bytes()
	subCh := make(chan common.Hash)
	sub, err := subClient.SubscribePendingTransactions(context.Background(), subCh)
	if err != nil {
		log.Fatalf("subscribe err: %v", err)
	}
	cnt, pendingCnt, wrongCnt, notFoundCnt := 0, 0, 0, 0
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("subcribe chan err: %v", err)
		case ev := <-subCh:
			cnt++
			fmt.Println(ev)
			tx, pending, err := client.TransactionByHash(context.Background(), ev)
			if err != nil {
				log.Printf("transcation by hash err: %v\n", err)
				notFoundCnt++
				continue
			}
			if !pending {
				pendingCnt++
				continue
			}
			address := tx.To()
			if address == nil || !bytes.Equal(address.Bytes(), exchangeContract) {
				wrongCnt++
				continue
			}
			// TODO: do smth here
		}
	}
}

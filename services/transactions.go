package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gin-gonic/gin"
)

// TransactionsResponse is returned by GetLatestBlockTransactions
type TransactionsResponse struct {
	Header       *types.Header
	Transactions []*types.Transaction
}

func weiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}

func getEthClient() *ethclient.Client {
	client, err := ethclient.Dial("replace_me")
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func getHeader(ctx context.Context, client *ethclient.Client) *types.Header {
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	return header
}

func getLatestBlock(ctx context.Context, client *ethclient.Client) uint64 {
	latestBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return latestBlockNumber
}

func inspectBlock(ctx context.Context, client *ethclient.Client) *types.Block {
	header := getHeader(ctx, client)
	blockNumber := big.NewInt(header.Number.Int64())
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(blockNumber, getLatestBlock(ctx, client))
	fmt.Println("Block Number:", block.Number().String())
	fmt.Println("Block Time:", block.Time())
	fmt.Println("Block Difficulty:", block.Difficulty().Uint64())
	fmt.Println("Block Hash:", block.Hash().Hex())
	fmt.Println("Block Transactions:", len(block.Transactions()), block.Transactions().Len())

	count, err := client.TransactionCount(ctx, block.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Count:", count)
	for _, tx := range block.Transactions() {
		fmt.Println(tx.Hash().Hex())
	}
	return block
}

func getMinerFromHeaderBlock(block *types.Block) map[string]interface{} {
	var headerInfo map[string]interface{}
	jsonByte, err := block.Header().MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(jsonByte), &headerInfo)
	return headerInfo
}

func getBalanceFromAccount(client *ethclient.Client, account common.Address) (balance *big.Int) {
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	return balance
}

// GetLatestBlockTransactions responds with the latest block with all transactions as JSON.
func GetLatestBlockTransactions(c *gin.Context) {
	ctx := context.Background()
	client := getEthClient()
	block := inspectBlock(ctx, client)
	minerInfo := getMinerFromHeaderBlock(block)
	account := common.HexToAddress(fmt.Sprintf("%v", minerInfo["miner"]))
	minerBalance := getBalanceFromAccount(client, account)
	fmt.Println("Miner account:", account, ", balance: ", weiToEther(minerBalance))
	m := TransactionsResponse{block.Header(), block.Body().Transactions}
	empJSON, err := json.MarshalIndent(block.Header(), "", "  ")
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf("%s\n", string(empJSON))
	c.IndentedJSON(http.StatusOK, m)
}

// StreamBlockTransactions streams block mined events as JSON
func StreamBlockTransactions(c *gin.Context) {
	ctx := context.Background()
	headers := make(chan *types.Header)
	chanStream := make(chan TransactionsResponse)
	client := getEthClient()
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer close(chanStream)
		for {
			select {
			case err := <-sub.Err():
				log.Fatal(err)
			case header := <-headers:

				block, err := client.BlockByHash(ctx, header.Hash())
				if err != nil {
					log.Fatal(err)
				}
				headerInfo := getMinerFromHeaderBlock(block)
				fmt.Println("Block Number", block.Number().Uint64())
				fmt.Println("Block Hash:", header.Hash().Hex())
				fmt.Println("Block Miner", headerInfo["miner"])
				fmt.Println("Block Transactions:", len(block.Transactions()))
				fmt.Println("------------------------------------------------------------------------------")
				m := TransactionsResponse{block.Header(), block.Body().Transactions}
				chanStream <- m
			}
		}
	}()
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-chanStream; ok {
			c.IndentedJSON(http.StatusOK, msg)
			return true
		}
		return false
	})
}

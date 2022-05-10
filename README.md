# eth-streamer

Go API with the GIN framework to live stream ethereum transactions

Simple project I made to learn about Goroutines

## Setup

Edit getEthClient in [transactions](/services/transactions.go) => replace rawurl parameter with something like `wss://mainnet.infura.io/ws/v3/xxxxxxxxxxx`

> `go mod tidy && go install`

### Dev

> `GIN_MODE=debug go run .`

### Prod

> `go build . && GIN_MODE=release ./eth-streamer.com`

### Trigger stream

> ```
> curl -s -D "/dev/stderr" http://localhost:8080/tx-start \
>  --include \
>  --header "Content-Type: application/json"
> ```

### Get latest TX

> ```
> curl -D "/dev/stderr" http://localhost:8080/transactions`
> ```
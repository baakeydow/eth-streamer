package main

import (
	ethStreamerRoutes "eth-streamer.com/routes"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
)

func main() {
	gin.DefaultWriter = colorable.NewColorableStderr()
	router := gin.Default()
	ethStreamerRoutes.SetTransactionsRoutes(router)
	router.Run("localhost:8080")
}

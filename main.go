package main

import (
	"balance-server/server"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onatm/clockwerk"
)

var (
	router = gin.Default()
	db     = server.NewDatabase()
	accRep = server.NewAccountRepository(db)
	acc    = server.NewAccountController(accRep)
)

func main() {
	defer db.Close()

	txVR := server.NewTransactionViewsRefresher(db)
	c := clockwerk.New()
	c.Every(3 * time.Minute).Do(txVR)

	c.Start()

	router.NoRoute(server.NoRoute)
	router.POST(server.URL_TRANSACTION, acc.Transaction)
	router.POST(server.URL_TRANSFER, acc.Transfer)
	router.GET(server.URL_BALANCE, acc.Balance)
	router.GET(server.URL_TRANSACTIONS, acc.Transactions)
	router.Run()
}

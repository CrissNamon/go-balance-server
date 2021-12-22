package main

import (
	"balance-server/server"

	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
	db     = server.NewDatabase()
	accRep = server.NewAccountRepository(db)
	acc    = server.NewAccountController(accRep)
)

func main() {
	defer db.Close()

	router.NoRoute(server.NoRoute)
	router.POST(server.URL_TRANSACTION, acc.Transaction)
	router.POST(server.URL_TRANSFER, acc.Transfer)
	router.GET(server.URL_BALANCE, acc.Balance)
	router.GET(server.URL_TRANSACTIONS, acc.Transactions)
	router.Run()
}

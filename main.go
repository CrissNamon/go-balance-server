// sandbox project main.go
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
	router.GET("/transaction", acc.Transaction)
	router.GET("/transfer", acc.Transfer)
	router.GET("/balance", acc.Balance)
	router.GET("/transactions", acc.Transactions)
	router.Run()
}

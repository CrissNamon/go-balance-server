package tests

import (
	"balance-server/server"
	"fmt"
	"testing"
)

func TestConnection(t *testing.T) {
	db := server.NewDatabase()
	if db.Conn == nil {
		t.Error("Connection error")
	}
	fmt.Println("Database connected successfully")
}

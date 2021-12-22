package tests

import (
	"balance-server/server"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	testDb  = NewTestDatabase()
	testRep = server.NewAccountRepository(testDb)
)

func TestConnection(t *testing.T) {
	if testDb.Conn == nil {
		t.Error("Database connection error")
	}
	t.Log("Connected to database: " + testDb.Conn.Config().ConnString())
}

func TestTransactionWrongRequest(t *testing.T) {
	_, err := testDb.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		return (*tx).Exec(testDb.Ctx, "SELECT FROM transaction;")
	})
	if err == nil {
		t.Error("Error expected, but hasn't been thrown")
	}
}

func TestTransactionBalanceNotExistingUser(t *testing.T) {
	exp := float64(0)
	bal, err := testDb.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		var curBal float64
		err := (*tx).QueryRow(testDb.Ctx, server.SELECT_CURRENT_BALANCE, 123125).Scan(&curBal)
		return curBal, err
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if bal != exp {
		t.Error("Expected balance: ", exp, ", but got: ", bal)
	}
}

func NewTestDatabase() *server.Database {
	testDb := server.Database{}
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		panic(err)
	}
	testDb.Ctx = context.Background()
	testDb.Conn = conn
	return &testDb
}

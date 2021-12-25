package tests

import (
	"balance-server/server"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

var (
	testDb  = NewTestDatabase()
	testRep = server.NewAccountRepository(testDb)
)

func TestConnection(t *testing.T) {
	assert.NotNil(t, testDb.Conn, "Database connection error")
	t.Log("Connected to database: " + testDb.Conn.Config().ConnString())
}

func TestTransactionWrongRequest(t *testing.T) {
	_, err := testDb.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		return (*tx).Exec(testDb.Ctx, "SELECT FROM transaction;")
	})
	assert.NotNil(t, err, "Error expected, but hasn't been thrown")
}

func TestTransactionBalanceNotExistingUser(t *testing.T) {
	_, err := testDb.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		var curBal *float64
		err := (*tx).QueryRow(testDb.Ctx, server.SELECT_CURRENT_BALANCE, 123125).Scan(&curBal)
		return curBal, err
	})
	assert.Nil(t, err)
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

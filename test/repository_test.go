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
	exp := float64(0)
	bal, err := testDb.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		var curBal float64
		err := (*tx).QueryRow(testDb.Ctx, server.SELECT_CURRENT_BALANCE, 123125).Scan(&curBal)
		return curBal, err
	})
	assert.Nil(t, err)
	assert.Equal(t, exp, bal, "Expected balance: ", exp, ", but got: ", bal)
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

type MockAccountRepository struct {
	executeTransactionFunc          func(trxData server.TransactionData, oCode int) error
	executeOperationFunc            func(trxData server.TransactionData) error
	getBalanceFunc                  func(dt server.BalanceData) (float64, error)
	executeTransferFunc             func(tData server.TransferData) error
	getTransactionsSortedByDateFunc func(trxData server.TransactionsListData) ([]map[string]interface{}, error)
	getTransactionsSortedBySumFunc  func(trxData server.TransactionsListData) ([]map[string]interface{}, error)
}

func NewMockRepository() *MockAccountRepository {
	return &MockAccountRepository{}
}

func (rep *MockAccountRepository) ExecuteTransaction(trxData server.TransactionData, oCode int) error {
	return rep.executeTransactionFunc(trxData, oCode)
}

func (rep *MockAccountRepository) ExecuteOperation(trxData server.TransactionData) error {
	if trxData.Sum > 0 {
		return rep.executeTransactionFunc(trxData, server.OPERATION_INCOME_CODE)
	} else {
		return rep.executeTransactionFunc(trxData, server.OPERATION_OUTCOME_CODE)
	}
}

func (rep *MockAccountRepository) GetBalance(dt server.BalanceData) (float64, error) {
	return rep.getBalanceFunc(dt)
}

func (rep *MockAccountRepository) ExecuteTransfer(tData server.TransferData) error {
	return rep.executeTransferFunc(tData)
}

func (rep *MockAccountRepository) GetTransactionsSortedByDate(trxData server.TransactionsListData) ([]map[string]interface{}, error) {
	return rep.getTransactionsSortedByDateFunc(trxData)
}

func (rep *MockAccountRepository) GetTransactionsSortedBySum(trxData server.TransactionsListData) ([]map[string]interface{}, error) {
	return rep.getTransactionsSortedBySumFunc(trxData)
}

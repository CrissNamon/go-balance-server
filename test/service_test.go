package tests

import (
	"balance-server/server"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

type MockAccountRepository struct {
	executeTransactionFunc          func(trxData server.TransactionData, oCode int) error
	executeOperationFunc            func(trxData server.TransactionData) error
	getBalanceFunc                  func(dt server.BalanceData) (float64, error)
	executeTransferFunc             func(tData server.TransferData) error
	getTransactionsSortedByDateFunc func(trxData server.TransactionsListData) (pgx.Rows, error)
	getTransactionsSortedBySumFunc  func(trxData server.TransactionsListData) (pgx.Rows, error)
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

func (rep *MockAccountRepository) GetTransactionsSortedByDate(trxData server.TransactionsListData) (pgx.Rows, error) {
	return rep.getTransactionsSortedByDateFunc(trxData)
}

func (rep *MockAccountRepository) GetTransactionsSortedBySum(trxData server.TransactionsListData) (pgx.Rows, error) {
	return rep.getTransactionsSortedBySumFunc(trxData)
}

func TestGetBalanceWrongCurrencyCode(t *testing.T) {
	rep := &MockAccountRepository{
		getBalanceFunc: func(dt server.BalanceData) (float64, error) {
			return 0, nil
		},
	}
	srv := server.NewAccountService(rep)
	_, err := srv.GetUserBalance(&server.BalanceData{1, "WRONG"})
	assert.NotNil(t, err, "Expected error, but hasn't been thrown")
	switch e := (err).(type) {
	case *server.OperationError:
		assert.Equal(t, server.ERROR_BALANCE_WRONG_CURRENCY_CODE, e.Code, "Expected code: ", server.ACCOUNT_OPERATION_STATUS[server.ERROR_BALANCE_WRONG_CURRENCY_CODE], ", but got: ", e.Code)
	default:
		t.Error("Expected OperationError, but got: ", e)
	}
}

func TestGetUserTransactionsWrongSort(t *testing.T) {
	rep := &MockAccountRepository{}
	srv := server.NewAccountService(rep)
	data := &server.TransactionsListData{Sort: "wrong"}
	_, err := srv.GetUserTransactions(data)
	assert.NotNil(t, err, "Expected error, but hasn't been thrown")
	switch e := (err).(type) {
	case *server.OperationError:
		assert.Equal(t, server.ERROR_TRANSACTIONS_WRONG_SORT, e.GetCode(), "Expected: ", server.ERROR_TRANSACTIONS_WRONG_SORT, ", but got: ", e.GetCode())
	default:
		t.Error("Expected OperationError, other has been thrown")
	}
}

func TestGetUserTransactionsWrongPage(t *testing.T) {
	rep := &MockAccountRepository{
		getTransactionsSortedByDateFunc: func(trxData server.TransactionsListData) (pgx.Rows, error) {
			return testDb.Conn.Query(testDb.GetCtx(), "SELECT * FROM transactions WHERE account = 9999")
		},
	}
	srv := server.NewAccountService(rep)
	data := &server.TransactionsListData{Page: 100}
	_, err := srv.GetUserTransactions(data)
	assert.NotNil(t, err, "Expected error, but hasn't been thrown")
	switch e := (err).(type) {
	case *server.OperationError:
		assert.Equal(t, server.ERROR_TRANSACTIONS_WRONG_PAGE, e.GetCode(), "Expected: ", server.ERROR_TRANSACTIONS_WRONG_PAGE, ", but got: ", e.GetCode())
	default:
		t.Error("Expected OperationError, other has been thrown")
	}
}

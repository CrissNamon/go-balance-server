package tests

import (
	"balance-server/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, server.STATUS_CODE_WRONG_CURRENCY_CODE, e.Code, "Expected code: ", server.STATUS_CODE_WRONG_CURRENCY_CODE, ", but got: ", e.Code)
	default:
		t.Error("Expected OperationError, but got: ", e)
	}
}

package tests

import (
	"balance-server/server"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	URL_HOST string = "http://localhost:8080"

	DB_INIT_QUERY string = "DELETE FROM transactions;"

	STATUS_ANY string = "STATUS_ANY_VAL"
)

type TestTable struct {
	Status   int
	Message  interface{}
	HttpCode int
}

var (
	rec    = httptest.NewRecorder()
	c, _   = gin.CreateTestContext(rec)
	acc    = server.NewAccountController(testRep)
	router = gin.New()
)

func TestMain(m *testing.M) {
	router.POST(server.URL_TRANSACTION, acc.Transaction)
	router.POST(server.URL_TRANSFER, acc.Transfer)
	router.GET(server.URL_BALANCE, acc.Balance)
	router.GET(server.URL_TRANSACTIONS, acc.Transactions)
	testDb.Conn.Query(testDb.Ctx, DB_INIT_QUERY)
	m.Run()
}

func TestBalanceNotExistingUser(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_OK, float64(0), 200}
	res := TestTable{}
	d := server.BalanceRequest{Id: 1, Cur: "RUB"}
	makeRequest(t, "GET", server.URL_BALANCE, &d, &res)
	httpTest(t, &res, &exp)
}

func TestBalanceWrongUser(t *testing.T) {
	s := fmt.Sprintf(server.BAD_REQUEST_BINDING, server.STATUS_WRONG_ID)
	exp := TestTable{server.STATUS_CODE_WRONG_REQUEST, s, 400}
	res := TestTable{}
	d := server.BalanceRequest{Id: -199, Cur: "RUB"}
	makeRequest(t, "GET", server.URL_BALANCE, &d, &res)
	httpTest(t, &res, &exp)
}

func TestBalanceWrongCurrencyCode(t *testing.T) {
	exp := TestTable{server.ERROR_BALANCE_WRONG_CURRENCY_CODE, server.AccountExpectedResult.GetStatus(server.ERROR_BALANCE_WRONG_CURRENCY_CODE), 400}
	res := TestTable{}
	d := server.BalanceRequest{Id: 1, Cur: "WRONG_CURRENCY"}
	makeRequest(t, "GET", server.URL_BALANCE, &d, &res)
	httpTest(t, &res, &exp)
	d.Cur = "AAA"
	makeRequest(t, "GET", server.URL_BALANCE, &d, &res)
	httpTest(t, &res, &exp)
}

func TestTransactionIncome(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_OK, server.STATUS_TRANSACTION_COMPLETED, 200}
	res := TestTable{}
	d := server.TransactionRequest{Id: 1, Sum: 100, Desc: ""}
	makeRequest(t, "POST", server.URL_TRANSACTION, &d, &res)
	httpTest(t, &res, &exp)
}

func TestTransactionOutcomeNoMoney(t *testing.T) {
	exp := TestTable{server.ERROR_NOT_ENOUGH_MONEY, server.ACCOUNT_OPERATION_STATUS[server.ERROR_NOT_ENOUGH_MONEY], 200}
	res := TestTable{}
	d := server.TransactionRequest{Id: 1, Sum: -10000, Desc: ""}
	makeRequest(t, "POST", server.URL_TRANSACTION, &d, &res)
	httpTest(t, &res, &exp)
}

func TestTransactionZeroSum(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_WRONG_REQUEST, STATUS_ANY, 400}
	res := TestTable{}
	d := server.TransactionRequest{Id: 1, Sum: 0, Desc: ""}
	makeRequest(t, "POST", server.URL_TRANSACTION, &d, &res)
	httpTest(t, &res, &exp)
}

func TestTransferSuccess(t *testing.T) {
	pD := server.TransactionRequest{Id: 1, Sum: 100, Desc: ""}
	makeRequest(t, "POST", server.URL_TRANSACTION, &pD, nil)
	exp := TestTable{server.STATUS_CODE_OK, server.STATUS_TRANSFER_COMPLETED, 200}
	res := TestTable{}
	d := server.SendRequest{From: 1, Sum: 100, To: 2}
	makeRequest(t, "POST", server.URL_TRANSFER, &d, &res)
	httpTest(t, &res, &exp)
}

func TestTransferEqualIds(t *testing.T) {
	s := fmt.Sprintf(server.BAD_REQUEST_BINDING, server.STATUS_WRONG_IDS_NOT_UNIQUE)
	exp := TestTable{server.STATUS_CODE_WRONG_REQUEST, s, 400}
	res := TestTable{}
	d := server.SendRequest{From: 2, Sum: 100, To: 2}
	makeRequest(t, "POST", server.URL_TRANSFER, &d, &res)
	httpTest(t, &res, &exp)
}

func makeRequest(t *testing.T, m string, path string, d interface{}, res *TestTable) {
	defer func() {
		rec = httptest.NewRecorder()
	}()
	data, err := json.Marshal(d)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(data))
	req, err := http.NewRequest(m, path, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(rec, req)
	if res != nil {
		var result map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&result)

		if err != nil {
			t.Error(err)
			return
		}
		res.HttpCode = rec.Code
		res.Status = int(result["status"].(float64))
		res.Message = result["data"]
	}
	return
}

func httpTest(t *testing.T, get *TestTable, want *TestTable) {
	assert.Equal(t, want.HttpCode, get.HttpCode, "Expexted HTTP status: ", want.HttpCode, ", but got: ", get.HttpCode)
	assert.Equal(t, want.Status, get.Status, "Expexted status: ", want.Status, ", but got: ", get.Status)
	if want.Message != STATUS_ANY {
		assert.Equal(t, want.Message, get.Message, "Expexted data: ", want.Message, ", but got: ", get.Message)
	}
}

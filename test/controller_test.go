package tests

import (
	"balance-server/server"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	URL_HOST string = "http://localhost:8080"

	DB_INIT_QUERY string = "DELETE FROM transactions;"
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
	router.GET(server.URL_TRANSACTION, acc.Transaction)
	router.GET(server.URL_TRANSFER, acc.Transfer)
	router.GET(server.URL_BALANCE, acc.Balance)
	router.GET(server.URL_TRANSACTIONS, acc.Transactions)
	testDb.Conn.Query(testDb.Ctx, DB_INIT_QUERY)
	m.Run()
}

func TestBalanceNotExistingUser(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_WRONG_REQUEST, "Wrong request data: id must be positive", 400}
	res := TestTable{}
	q := make(url.Values)
	q.Add("id", "wrong")
	makeRequest(t, server.URL_BALANCE, &q, &res)
	httpTest(t, &res, &exp)
}

func TestBalanceWrongUser(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_WRONG_REQUEST, "Wrong request data: id must be positive", 400}
	res := TestTable{}
	q := make(url.Values)
	q.Add("id", "wrong")
	makeRequest(t, server.URL_BALANCE, &q, &res)
	httpTest(t, &res, &exp)
	q.Set("id", "-199")
	makeRequest(t, server.URL_BALANCE, &q, &res)
	httpTest(t, &res, &exp)
}

func TestNewTransactionIncome(t *testing.T) {
	exp := TestTable{0, server.STATUS_TRANSACTION_COMPLETED, 200}
	res := TestTable{}
	q := make(url.Values)
	q.Add("id", "1")
	q.Add("sum", "100")
	makeRequest(t, server.URL_TRANSACTION, &q, &res)
	httpTest(t, &res, &exp)
}

func TestNewTransactionOutcomeNoMoney(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_NOT_ENOUGH_MONEY, server.STATUS_NOT_ENOUGHT_MONEY, 200}
	res := TestTable{}
	q := make(url.Values)
	q.Add("id", "1")
	q.Add("sum", "-1000")
	makeRequest(t, server.URL_TRANSACTION, &q, &res)
	httpTest(t, &res, &exp)
}

func makeRequest(t *testing.T, path string, q *url.Values, res *TestTable) {
	defer func() {
		rec = httptest.NewRecorder()
	}()
	req, _ := http.NewRequest("GET", path, nil)
	req.URL.RawQuery = q.Encode()
	router.ServeHTTP(rec, req)

	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)

	if err != nil {
		t.Error(err)
		return
	}
	res.HttpCode = rec.Code
	res.Status = int(result["status"].(float64))
	res.Message = result["data"]
	return
}

func httpTest(t *testing.T, get *TestTable, want *TestTable) {
	if get.HttpCode != want.HttpCode {
		t.Error("Expexted HTTP status: ", want.HttpCode, ", but got: ", get.HttpCode)
	}
	if get.Status != want.Status {
		t.Error("Expexted status: ", want.Status, ", but got: ", get.Status)
	}
	if get.Message != want.Message {
		t.Error("Expected data: ", want.Message, ", but got: ", get.Message)
	}
}

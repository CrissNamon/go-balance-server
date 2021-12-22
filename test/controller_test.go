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
	rec  = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	acc  = server.NewAccountController(testRep)
)

func TestMain(m *testing.M) {
	testDb.Conn.Query(testDb.Ctx, DB_INIT_QUERY)
	m.Run()
}

func TestBalanceNotExistingUser(t *testing.T) {
	exp := TestTable{0, float64(0), 200}
	q := make(url.Values)
	q.Add("id", "1")
	res := TestTable{}
	makeRequest(t, &q, acc.Balance, &res)
	httpTest(t, &res, &exp)
}

func TestNewTransactionIncome(t *testing.T) {
	exp := TestTable{0, server.STATUS_TRANSACTION_COMPLETED, 200}
	q := make(url.Values)
	q.Add("id", "1")
	q.Add("sum", "100")
	res := TestTable{}
	makeRequest(t, &q, acc.Transaction, &res)
	httpTest(t, &res, &exp)
}

func TestNewTransactionOutcomeNoMoney(t *testing.T) {
	exp := TestTable{server.STATUS_CODE_NOT_ENOUGH_MONEY, server.STATUS_NOT_ENOUGHT_MONEY, 200}
	q := make(url.Values)
	q.Add("id", "1")
	q.Add("sum", "-1000")
	res := TestTable{}
	makeRequest(t, &q, acc.Transaction, &res)
	httpTest(t, &res, &exp)
}

func makeRequest(t *testing.T, q *url.Values, m func(c *gin.Context), res *TestTable) {
	defer rec.Flush()

	req := &http.Request{URL: &url.URL{}}
	req.URL.RawQuery = q.Encode()
	c.Request = req
	m(c)
	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)

	if err != nil {
		t.Error(err)
		return
	}
	t.Log(result)
	res.HttpCode = rec.Result().StatusCode
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

package server

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	URL_BALANCE      string = "/balance"
	URL_TRANSACTION  string = "/transaction"
	URL_TRANSACTIONS string = "/transactions"
	URL_TRANSFER     string = "/transfer"

	STATUS_CODE_OK             int = 0
	STATUS_CODE_WRONG_REQUEST  int = 1
	STATUS_CODE_INTERNAL_ERROR int = 3

	STATUS_NOT_ENOUGHT_MONEY     string = "Not enought money"
	STATUS_TRANSACTION_COMPLETED string = "Transaction completed"
	STATUS_INTERNAL_ERROR        string = "Internal server error"
	STATUS_TRANSFER_COMPLETED    string = "Transfer completed"
	STATUS_WRONG_CURRENCY_CODE   string = "Wrong currency code"
	STATUS_WRONG_SORT            string = "Wrong sorting key"
	STATUS_WRONG_PAGE            string = "Wrong page number"

	PAGINATION_PAGE_SIZE int = 2
)

var (
	ACCOUNT_OPERATION_STATUS = map[int]string{
		ERROR_NOT_ENOUGH_MONEY:            STATUS_NOT_ENOUGHT_MONEY,
		ERROR_BALANCE_WRONG_CURRENCY_CODE: STATUS_WRONG_CURRENCY_CODE,
		ERROR_TRANSACTIONS_WRONG_PAGE:     STATUS_WRONG_PAGE,
		ERROR_TRANSACTIONS_WRONG_SORT:     STATUS_WRONG_SORT,
	}

	ACCOUNT_OPERATION_RESPONSE_CODE = map[int]int{
		ERROR_TRANSACTIONS_WRONG_PAGE:     400,
		ERROR_TRANSACTIONS_WRONG_SORT:     400,
		ERROR_BALANCE_WRONG_CURRENCY_CODE: 400,
	}
)

var (
	AccountExpectedResult = ExpectedResult{
		ACCOUNT_OPERATION_STATUS,
		ACCOUNT_OPERATION_RESPONSE_CODE,
	}
)

type TransactionRequest struct {
	Id   int     `form:"id" json:"id" binding:"required,numeric,gte=0"`
	Sum  float64 `form:"sum" json:"sum" binding:"required,numeric"`
	Desc string  `form:"desc" json:"desc"`
}

type SendRequest struct {
	From int     `form:"id" json:"id" binding:"required,numeric,gte=0"`
	Sum  float64 `form:"sum" json:"sum" binding:"required,numeric,gt=0"`
	To   int     `form:"to" json:"to" binding:"required,numeric,gte=0"`
}

type BalanceRequest struct {
	Id  int    `form:"id" json:"id" binding:"required,numeric,gte=0"`
	Cur string `form:"currency" json:"currency"`
}

type TransactionsRequest struct {
	Id   int    `form:"id" json:"id" binding:"required,numeric,gte=0`
	From int64  `form:"from" json:"from"`
	To   int64  `form:"to" json:"to"`
	Sort string `form:"sort" json:"sort"`
	Page int    `form:"page" json:"page" binding:"gte=0"`
}

type AccountController struct {
	accSrv *AccountService
}

func NewAccountController(accRep AccountRepositoryI) *AccountController {
	s := NewAccountService(accRep)
	return &AccountController{s}
}

func (acc *AccountController) Transaction(c *gin.Context) {
	var trxReq TransactionRequest
	r := Result{c, STATUS_CODE_OK, STATUS_TRANSACTION_COMPLETED}
	if err := c.ShouldBindJSON(&trxReq); err != nil {
		r.BadRequest("id must be positive and sum must be greater than zero.")
		return
	}
	trxData := TransactionData{trxReq.Id, trxReq.Sum, trxReq.Desc}
	err := acc.accSrv.DoTransaction(&trxData)
	if err != nil {
		r.Err(&err, &AccountExpectedResult)
		return
	}
	r.Ok()
}

func (acc *AccountController) Transfer(c *gin.Context) {
	var sReq SendRequest
	r := Result{c, STATUS_CODE_OK, STATUS_TRANSFER_COMPLETED}
	if err := c.ShouldBindJSON(&sReq); err != nil {
		r.BadRequest("ids must be positive and sum must be greater than zero.")
		return
	}
	tData := TransferData{sReq.From, sReq.To, sReq.Sum}
	err := acc.accSrv.TransferMoney(&tData)
	if err != nil {
		r.Err(&err, &AccountExpectedResult)
		return
	}
	r.Ok()
}

func (acc *AccountController) Balance(c *gin.Context) {
	r := Result{c, STATUS_CODE_OK, 0}
	blncReq := BalanceRequest{0, BASE_CURRENCY}
	if err := c.ShouldBindJSON(&blncReq); err != nil {
		r.BadRequest(err.Error())
		return
	}
	bData := BalanceData{blncReq.Id, blncReq.Cur}
	curBal, err := acc.accSrv.GetUserBalance(&bData)
	if err != nil {
		r.Err(&err, &AccountExpectedResult)
		return
	}
	r.Give(curBal)
}

func (acc *AccountController) Transactions(c *gin.Context) {
	r := Result{c, STATUS_CODE_OK, map[string]interface{}{}}
	var trxsReq TransactionsRequest
	if err := c.ShouldBindJSON(&trxsReq); err != nil {
		r.BadRequest("ids must be positive")
		return
	}
	to := trxsReq.To
	if to == 0 {
		to = time.Now().Unix()
	}
	trxData := TransactionsListData{trxsReq.Id, trxsReq.From, to, trxsReq.Page, trxsReq.Sort}
	trxs, err := acc.accSrv.GetUserTransactions(&trxData)
	if err != nil {
		r.Err(&err, &AccountExpectedResult)
		return
	}
	r.Give(trxs)
}

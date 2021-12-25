package server

import (
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	ERROR_BALANCE_WRONG_CURRENCY_CODE int = 101
	ERROR_TRANSACTIONS_WRONG_SORT     int = 102
	ERROR_TRANSACTIONS_WRONG_PAGE     int = 103
	ERROR_NOT_ENOUGH_MONEY            int = 104
	ERROR_WRONG_USER_ID               int = 105
	ERROR_LOCK_TIMEOUT                int = 106
	ERROR_NO_BALANCE                  int = 107
)

type TransactionData struct {
	Id   int
	Sum  float64
	Desc string
}

type BalanceData struct {
	Id  int
	Cur string
}

type TransferData struct {
	From int
	To   int
	Sum  float64
}

type TransactionsListData struct {
	Id   int
	From int64
	To   int64
	Page int
	Sort string
}

type AccountService struct {
	accRep AccountRepositoryI
}

func NewAccountService(r AccountRepositoryI) *AccountService {
	return &AccountService{r}
}

func (s *AccountService) GetUserBalance(bData *BalanceData) (float64, error) {
	if len(bData.Cur) != 3 {
		return 0, &OperationError{ERROR_BALANCE_WRONG_CURRENCY_CODE}
	}
	curBal, err := s.accRep.GetBalance(*bData)
	if err != nil {
		return 0, ConvertError(err)
	}
	if bData.Cur != BASE_CURRENCY {
		rate, err := GetCurrencyRate(BASE_CURRENCY, (*bData).Cur)
		if err != nil {
			return 0, ConvertError(err)
		}
		curBal *= rate
	}
	return curBal, nil
}

func (s *AccountService) GetUserTransactions(trxData *TransactionsListData) (TransactionsData, error) {
	var trxs []map[string]interface{}
	var err error
	var last int
	var rows pgx.Rows
	switch trxData.Sort {
	case "date":
		rows, err = s.accRep.GetTransactionsSortedByDate(*trxData)
	case "sum":
		rows, err = s.accRep.GetTransactionsSortedBySum(*trxData)
	case "":
		rows, err = s.accRep.GetTransactionsSortedByDate(*trxData)
	default:
		return TransactionsData{}, &OperationError{ERROR_TRANSACTIONS_WRONG_SORT}
	}
	if err != nil {
		return TransactionsData{}, ConvertError(err)
	}
	last, trxs, err = s.transactionRowsToArray(&rows)
	if err != nil {
		return TransactionsData{}, ConvertError(err)
	}
	if len(trxs) == 0 && trxData.Page > 0 {
		return TransactionsData{}, &OperationError{ERROR_TRANSACTIONS_WRONG_PAGE}
	}
	return TransactionsData{last, trxs}, nil
}

func (s *AccountService) TransferMoney(tData *TransferData) error {
	err := s.accRep.ExecuteTransfer(*tData)
	if err != nil {
		return ConvertError(err)
	}
	return nil
}

func (s *AccountService) DoTransaction(tData *TransactionData) error {
	err := s.accRep.ExecuteOperation(*tData)
	if err != nil {
		return ConvertError(err)
	}
	return nil
}

func (s *AccountService) transactionRowsToArray(rows *pgx.Rows) (last int, trxs []map[string]interface{}, err error) {
	trxs = []map[string]interface{}{}
	var (
		sum       float64
		operation int
		date      int64
		desc      string
	)
	for (*rows).Next() {
		err = (*rows).Scan(&last, &sum, &operation, &date, &desc)
		trx := map[string]interface{}{
			"sum":       sum,
			"operation": operation,
			"date":      time.Unix(date, 0),
			"desc":      desc,
		}
		if err != nil {
			return 0, trxs, err
		}
		trxs = append(trxs, trx)
	}
	if err != nil {
		return 0, nil, err
	}
	l := len(trxs)
	if l > PAGINATION_PAGE_SIZE {
		trxs = trxs[:l-1]
	}
	if l <= PAGINATION_PAGE_SIZE {
		last = -1
	}
	return last, trxs, err
}

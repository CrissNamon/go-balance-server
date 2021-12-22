package server

import (
	"math"
)

const (
	ERROR_BALANCE_WRONG_CURRENCY_CODE int = 101
	ERROR_TRANSACTIONS_WRONG_SORT     int = 102
	ERROR_TRANSACTIONS_WRONG_PAGE     int = 103
	ERROR_NOT_ENOUGH_MONEY            int = 104
)

type AccountService struct {
	accRep AccountRepositoryI
}

func NewAccountService(r AccountRepositoryI) *AccountService {
	return &AccountService{r}
}

func (s *AccountService) GetUserBalance(bData *BalanceData) (float64, error) {
	curBal, err := s.accRep.GetBalance(*bData)
	if err != nil {
		return 0, ConvertError(err)
	}
	if len(bData.Cur) != 3 {
		return 0, &OperationError{ERROR_BALANCE_WRONG_CURRENCY_CODE}
	}
	if bData.Cur != BASE_CURRENCY {
		if rate, err := GetCurrencyRate(BASE_CURRENCY, (*bData).Cur); err != nil {
			return 0, ConvertError(err)
		} else {
			curBal *= rate
		}
	}
	return curBal, nil
}

func (s *AccountService) GetUserTransactions(trxData *TransactionsListData) ([]map[string]interface{}, error) {
	var trxs []map[string]interface{}
	var err error
	switch trxData.Sort {
	case "date":
		trxs, err = s.accRep.GetTransactionsSortedByDate(*trxData)
	case "sum":
		trxs, err = s.accRep.GetTransactionsSortedBySum(*trxData)
	case "":
		trxs, err = s.accRep.GetTransactionsSortedByDate(*trxData)
	default:
		return nil, &OperationError{ERROR_TRANSACTIONS_WRONG_SORT}
	}
	if err != nil {
		return nil, ConvertError(err)
	}
	if trxData.Page == 0 {
		return trxs, nil
	}
	l := len(trxs)
	pgs := int(math.Ceil(float64(l) / float64(PAGINATION_PAGE_SIZE)))
	if trxData.Page > pgs {
		return nil, &OperationError{ERROR_TRANSACTIONS_WRONG_PAGE}
	}
	start := (trxData.Page - 1) * PAGINATION_PAGE_SIZE
	end := trxData.Page * PAGINATION_PAGE_SIZE
	if end > l {
		end = l
	}
	return trxs[start:end], nil
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

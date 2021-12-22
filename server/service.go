package server

import (
	"math"
)

type AccountService struct {
	accRep *AccountRepository
}

func NewAccountService(r *AccountRepository) *AccountService {
	return &AccountService{r}
}

func (s *AccountService) GetUserBalance(bData *BalanceData) (float64, error) {
	curBal, err := s.accRep.getBalance(*bData)
	if err != nil {
		return 0, err
	}
	if len(bData.Cur) != 3 {
		return 0, &OperationError{STATUS_CODE_WRONG_CURRENCY_CODE}
	}
	if bData.Cur != BASE_CURRENCY {
		if rate, err := GetCurrencyRate(BASE_CURRENCY, (*bData).Cur); err != nil {
			return 0, err
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
		trxs, err = s.accRep.getTransactionsSortedByDate(*trxData)
	case "sum":
		trxs, err = s.accRep.getTransactionsSortedBySum(*trxData)
	case "":
		trxs, err = s.accRep.getTransactionsSortedByDate(*trxData)
	default:
		return nil, &OperationError{STATUS_CODE_WRONG_REQUEST}
	}
	if err != nil {
		return nil, err
	}
	if trxData.Page == 0 {
		return trxs, nil
	}
	l := len(trxs)
	pgs := int(math.Ceil(float64(l) / float64(PAGINATION_PAGE_SIZE)))
	if trxData.Page > pgs {
		return nil, &OperationError{STATUS_CODE_WRONG_REQUEST}
	}
	start := (trxData.Page - 1) * PAGINATION_PAGE_SIZE
	end := trxData.Page * PAGINATION_PAGE_SIZE
	if end > l {
		end = l
	}
	return trxs[start:end], nil
}

func (s *AccountService) TransferMoney(tData *TransferData) error {
	return s.accRep.executeTransfer(*tData)
}

func (s *AccountService) DoTransaction(tData *TransactionData) error {
	return s.accRep.executeOperation(*tData)
}

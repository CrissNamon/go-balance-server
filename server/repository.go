package server

import (
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	SELECT_CURRENT_BALANCE   string = "SELECT COALESCE(SUM(sum), 0) FROM transactions WHERE account = $1"
	COUNT_TRANSACTIONS       string = "SELECT COUNT(*) FROM transactions WHERE account = $1"
	CREATE_TRANSACTION       string = "INSERT INTO transactions(account, sum, operation, description) VALUES($1, $2, $3, $4)"
	GET_TRANSACTIONS_FROM_TO string = "SELECT sum, operation, date, description FROM transactions WHERE account = $1 AND date >= $2 AND date <= $3"

	SORT_TRANSACTIONS_DATE string = "ORDER BY date DESC"
	SORT_TRANSACTIONS_SUM  string = "ORDER BY sum DESC"

	OPERATION_INCOME_CODE   int = 0
	OPERATION_OUTCOME_CODE  int = 1
	OPERATION_TRANSFER_CODE int = 2

	OPERATION_TRANSFER_DESC string = "Transfer to user %d from user %d"
)

type TransactionData struct {
	Id  int
	Sum float64
}

type BalanceData struct {
	Id int
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
}

type AccountRepository struct {
	db *Database
}

func NewAccountRepository(db *Database) *AccountRepository {
	return &AccountRepository{db}
}

func (rep *AccountRepository) executeTransaction(trxData TransactionData, oCode int, desc string) error {
	_, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		var curBal float64
		err := (*tx).QueryRow(rep.db.ctx, SELECT_CURRENT_BALANCE, trxData.Id).Scan(&curBal)
		if err != nil {
			return nil, err
		}
		if trxData.Sum < 0 && math.Abs(curBal) < math.Abs(trxData.Sum) {
			return nil, &OperationError{STATUS_CODE_NOT_ENOUGH_MONEY}
		}
		_, err = (*tx).Exec(rep.db.ctx, CREATE_TRANSACTION, trxData.Id, trxData.Sum, oCode, desc)
		return nil, err
	})
	return err
}

func (rep *AccountRepository) executeOperation(trxData TransactionData, desc string) error {
	if trxData.Sum > 0 {
		return rep.executeTransaction(trxData, OPERATION_INCOME_CODE, desc)
	} else {
		return rep.executeTransaction(trxData, OPERATION_OUTCOME_CODE, desc)
	}
}

func (rep *AccountRepository) getBalance(dt BalanceData) (float64, error) {
	var curBal float64
	_, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		err := (*tx).QueryRow(rep.db.ctx, SELECT_CURRENT_BALANCE, dt.Id).Scan(&curBal)
		return nil, err
	})
	return curBal, err
}

func (rep *AccountRepository) executeTransfer(tData TransferData) error {
	trxData := TransactionData{tData.From, -tData.Sum}
	desc := fmt.Sprintf(OPERATION_TRANSFER_DESC, tData.To, tData.From)
	err := rep.executeTransaction(trxData, OPERATION_TRANSFER_CODE, desc)
	if err != nil {
		return err
	}
	trxData = TransactionData{tData.To, tData.Sum}
	err = rep.executeTransaction(trxData, OPERATION_TRANSFER_CODE, desc)
	return err
}

func (rep *AccountRepository) getTransactionsWithSort(trxData TransactionsListData, sort string) ([]map[string]interface{}, error) {
	qry := GET_TRANSACTIONS_FROM_TO + " " + sort
	trxs, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		rows, err := (*tx).Query(rep.db.ctx, qry, trxData.Id, trxData.From, trxData.To)
		if err != nil {
			return nil, err
		}
		return transactionRowsToArray(&rows)
	})
	return trxs.([]map[string]interface{}), err
}

func (rep *AccountRepository) getTransactionsSortedByDate(trxData TransactionsListData) ([]map[string]interface{}, error) {
	return rep.getTransactionsWithSort(trxData, SORT_TRANSACTIONS_DATE)
}

func (rep *AccountRepository) getTransactionsSortedBySum(trxData TransactionsListData) ([]map[string]interface{}, error) {
	return rep.getTransactionsWithSort(trxData, SORT_TRANSACTIONS_SUM)
}

func transactionRowsToArray(rows *pgx.Rows) (trxs []map[string]interface{}, err error) {
	trxs = []map[string]interface{}{}
	for (*rows).Next() {
		var (
			sum       float64
			operation int
			date      int64
			desc      string
		)
		err = (*rows).Scan(&sum, &operation, &date, &desc)
		trx := map[string]interface{}{
			"sum":       sum,
			"operation": operation,
			"date":      time.Unix(date, 0),
			"desc":      desc,
		}
		if err != nil {
			return trxs, err
		}
		trxs = append(trxs, trx)
	}
	return trxs, err
}

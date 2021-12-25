package server

import (
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	SELECT_ADVISORY_LOCK                            string = "SELECT pg_advisory_xact_lock($1, $2)"
	SET_LOCK_TIMEOUT                                string = "SET LOCAL lock_timeout = '10s'"
	SELECT_CURRENT_BALANCE                          string = "SELECT SUM(sum) FROM transactions WHERE account = $1"
	SELECT_CURRENT_BALANCE_COALESCE                 string = "SELECT COALESCE(SUM(sum), 0) FROM transactions WHERE account = $1"
	COUNT_TRANSACTIONS                              string = "SELECT COUNT(*) FROM transactions WHERE account = $1"
	CREATE_TRANSACTION                              string = "INSERT INTO transactions(account, sum, operation, description) VALUES($1, $2, $3, $4)"
	GET_TRANSACTIONS_FROM_TO_ORDERED_DATE_FIRSTPAGE string = "SELECT id, sum, operation, date, description FROM transactions WHERE account = $1 AND date >= $2 AND date <= $3 ORDER BY date DESC LIMIT $4"
	GET_TRANSACTIONS_FROM_TO_ORDERED_DATE           string = "SELECT id, sum, operation, date, description FROM transactions WHERE account = $1 AND date >= $2 AND date <= $3 AND id <= $4 ORDER BY date DESC LIMIT $5"
	GET_TRANSACTIONS_FROM_TO_ORDERED_SUM            string = "SELECT transactions.id, sum, operation, date, description FROM transactions_sum_order INNER JOIN transactions ON transactions_sum_order.id = transactions.id WHERE transactions.account = $1 AND transactions.date >= $2 AND transactions.date <= $3 AND transactions_sum_order.pager >= $4 ORDER BY pager ASC LIMIT $5"
	UPDATE_ORDERED_SUM_VIEW                         string = "REFRESH MATERIALIZED VIEW CONCURRENTLY transactions_sum_order"

	OPERATION_INCOME_CODE  int = 0
	OPERATION_OUTCOME_CODE int = 1

	OPERATION_TRANSFER_DESC string = "Transfer to user %d from user %d"
)

var (
	LOCKED_OPERATIONS = map[int]bool{
		OPERATION_OUTCOME_CODE: true,
	}
)

type TransactionViewsRefresher struct {
	db DatabaseI
}

func NewTransactionViewsRefresher(db DatabaseI) *TransactionViewsRefresher {
	return &TransactionViewsRefresher{db}
}

func (r *TransactionViewsRefresher) Run() {
	fmt.Println("Updating DB transaction_sum_order view")
	r.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		_, err := (*tx).Exec(r.db.GetCtx(), UPDATE_ORDERED_SUM_VIEW)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
}

type AccountRepositoryI interface {
	ExecuteTransaction(trxData TransactionData, oCode int) error
	ExecuteOperation(trxData TransactionData) error
	GetBalance(dt BalanceData) (float64, error)
	ExecuteTransfer(tData TransferData) error
	GetTransactionsSortedByDate(trxData TransactionsListData) (int, []map[string]interface{}, error)
	GetTransactionsSortedBySum(trxData TransactionsListData) (int, []map[string]interface{}, error)
}

type AccountRepository struct {
	db DatabaseI
}

func NewAccountRepository(db DatabaseI) *AccountRepository {
	return &AccountRepository{db}
}

func (rep *AccountRepository) ExecuteTransaction(trxData TransactionData, oCode int) error {
	_, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		var err error
		if l := rep.shouldBeLocked(oCode); l {
			(*tx).Exec(rep.db.GetCtx(), SET_LOCK_TIMEOUT)
			_, err = (*tx).Exec(rep.db.GetCtx(), SELECT_ADVISORY_LOCK, trxData.Id, oCode)
			if err != nil {
				return nil, &OperationError{ERROR_LOCK_TIMEOUT}
			}
		}
		var curBal float64
		err = (*tx).QueryRow(rep.db.GetCtx(), SELECT_CURRENT_BALANCE_COALESCE, trxData.Id).Scan(&curBal)
		if err != nil {
			return nil, err
		}
		if trxData.Sum < 0 && math.Abs(curBal) < math.Abs(trxData.Sum) {
			return nil, &OperationError{ERROR_NOT_ENOUGH_MONEY}
		}
		_, err = (*tx).Exec(rep.db.GetCtx(), CREATE_TRANSACTION, trxData.Id, trxData.Sum, oCode, trxData.Desc)
		return nil, err
	})
	return err
}

func (rep *AccountRepository) ExecuteOperation(trxData TransactionData) error {
	if trxData.Sum > 0 {
		return rep.ExecuteTransaction(trxData, OPERATION_INCOME_CODE)
	} else {
		return rep.ExecuteTransaction(trxData, OPERATION_OUTCOME_CODE)
	}
}

func (rep *AccountRepository) GetBalance(dt BalanceData) (float64, error) {
	var curBal *float64
	_, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		err := (*tx).QueryRow(rep.db.GetCtx(), SELECT_CURRENT_BALANCE, dt.Id).Scan(&curBal)
		if err != nil {
			fmt.Println(err.Error())
		}
		return nil, err
	})
	if curBal == nil {
		return 0, &OperationError{ERROR_NO_BALANCE}
	}
	return *curBal, err
}

func (rep *AccountRepository) ExecuteTransfer(tData TransferData) error {
	desc := fmt.Sprintf(OPERATION_TRANSFER_DESC, tData.To, tData.From)
	trxData := TransactionData{tData.From, -tData.Sum, desc}
	err := rep.ExecuteTransaction(trxData, OPERATION_OUTCOME_CODE)
	if err != nil {
		return err
	}
	trxData = TransactionData{tData.To, tData.Sum, desc}
	err = rep.ExecuteTransaction(trxData, OPERATION_INCOME_CODE)
	return err
}

func (rep *AccountRepository) getTransactions(qry string, args ...interface{}) (pgx.Rows, error) {
	rows, err := rep.db.ExecuteInTransaction(func(tx *pgx.Tx) (interface{}, error) {
		return (*tx).Query(rep.db.GetCtx(), qry, args...)

	})
	if err != nil {
		return nil, err
	}
	return rows.(pgx.Rows), nil
}

func (rep *AccountRepository) GetTransactionsSortedByDate(trxData TransactionsListData) (int, []map[string]interface{}, error) {
	var rows pgx.Rows
	var err error
	if trxData.Page == 0 {
		rows, err = rep.getTransactions(GET_TRANSACTIONS_FROM_TO_ORDERED_DATE_FIRSTPAGE, trxData.Id, trxData.From, trxData.To, PAGINATION_PAGE_SIZE+1)
	} else {
		rows, err = rep.getTransactions(GET_TRANSACTIONS_FROM_TO_ORDERED_DATE, trxData.Id, trxData.From, trxData.To, trxData.Page, PAGINATION_PAGE_SIZE+1)
	}
	if err != nil {
		return 0, nil, err
	}
	last, trxs, err := rep.transactionRowsToArray(rows.(pgx.Rows))
	l := len(trxs)
	if l > 0 {
		trxs = trxs[:l-1]
	}
	if l <= PAGINATION_PAGE_SIZE {
		last = -1
	}
	return last, trxs, err
}

func (rep *AccountRepository) GetTransactionsSortedBySum(trxData TransactionsListData) (int, []map[string]interface{}, error) {
	rows, err := rep.getTransactions(GET_TRANSACTIONS_FROM_TO_ORDERED_SUM, trxData.Id, trxData.From, trxData.To, trxData.Page, PAGINATION_PAGE_SIZE)
	if err != nil {
		return 0, nil, err
	}
	last, trxs, err := rep.transactionRowsToArray(rows.(pgx.Rows))
	if last == trxData.Page {
		last = -1
	} else {
		last++
	}
	return last, trxs, err
}

func (rep *AccountRepository) transactionRowsToArray(rows pgx.Rows) (last int, trxs []map[string]interface{}, err error) {
	trxs = []map[string]interface{}{}
	var (
		sum       float64
		operation int
		date      int64
		desc      string
	)
	for rows.Next() {
		err = rows.Scan(&last, &sum, &operation, &date, &desc)
		trx := map[string]interface{}{
			"id":        last,
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
	return last, trxs, err
}

func (rep *AccountRepository) shouldBeLocked(oCode int) bool {
	r, ok := LOCKED_OPERATIONS[oCode]
	if !ok {
		return false
	}
	return r
}

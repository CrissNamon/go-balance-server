package server

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	Conn *pgxpool.Pool
	Ctx  context.Context
}

func NewDatabase() *Database {
	db := Database{}
	db.Open()
	return &db
}

func (db *Database) Open() {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Print("Error on db connection: ")
		fmt.Println(err.Error())
		panic(err)
	}
	db.Ctx = context.Background()
	db.Conn = conn
	fmt.Println("Connected to database")
}

func (db *Database) Close() {
	db.Conn.Close()
	fmt.Println("Disconnected from database")
}

func (db *Database) ExecuteInTransaction(actn func(tx *pgx.Tx) (interface{}, error)) (interface{}, error) {
	tx, err := db.Conn.BeginTx(db.Ctx, pgx.TxOptions{})
	defer func() {
		if err != nil {
			tx.Rollback(db.Ctx)
		} else {
			tx.Commit(db.Ctx)
		}
	}()
	if err != nil {
		return nil, err
	}
	return actn(&tx)
}

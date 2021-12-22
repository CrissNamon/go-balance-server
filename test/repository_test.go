package tests

import (
	"balance-server/server"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	testDb  = NewTestDatabase()
	testRep = server.NewAccountRepository(testDb)
)

func TestConnection(t *testing.T) {
	if testDb.Conn == nil {
		t.Error("Database connection error")
	}
	t.Log("Connected to database: " + testDb.Conn.Config().ConnString())
}

func NewTestDatabase() *server.Database {
	testDb := server.Database{}
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		panic(err)
	}
	testDb.Ctx = context.Background()
	testDb.Conn = conn
	return &testDb
}

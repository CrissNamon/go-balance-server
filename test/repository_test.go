package tests

import (
	"balance-server/server"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	db  = NewTestDatabase()
	rep = server.NewAccountRepository(db)
)

func TestConnection(t *testing.T) {
	if db.Conn == nil {
		t.Error("Database connection error")
	}
	t.Log("Connected to database: " + db.Conn.Config().ConnString())
}

func NewTestDatabase() *server.Database {
	db := server.Database{}
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		panic(err)
	}
	db.Ctx = context.Background()
	db.Conn = conn
	return &db
}

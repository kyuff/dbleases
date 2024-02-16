package tests

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kyuff/dbleases/internal/assert"
)

func Connect(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:5430/%s",
		"lease",
		"lease",
		"localhost",
		"lease",
	)
	db, err := sql.Open("pgx", dsn)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, db.Ping()) {
		t.FailNow()
	}

	return db
}

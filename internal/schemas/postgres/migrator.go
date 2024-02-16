package postgres

import (
	"context"
	"database/sql"

	"github.com/kyuff/dbleases/internal/hash"
	"github.com/kyuff/dbleases/internal/migrator"
)

type Migrator struct {
	sql        map[string]string
	names      tableNames
	migrations map[string]string
}

func (m *Migrator) Migrations() map[string]string {
	return m.migrations
}

func (m *Migrator) CreateSchema(ctx context.Context, db migrator.DB) error {
	_, err := db.ExecContext(ctx, m.sql[createSchema])
	return err
}

func (m *Migrator) CreateMigrationTable(ctx context.Context, db migrator.DB) error {
	_, err := db.ExecContext(ctx, m.sql[createMigrationTable])
	return err
}

func (m *Migrator) SelectMaxMigration(ctx context.Context, db migrator.DB) *sql.Row {
	return db.QueryRowContext(ctx, m.sql[selectMaxMigration])
}

func (m *Migrator) InsertMigrationRow(ctx context.Context, db migrator.DB, version uint32, fileName, sha string) error {
	_, err := db.ExecContext(ctx, m.sql[insertMigrationRow], version, fileName, sha)
	return err
}

func (m *Migrator) SelectLock(ctx context.Context, db migrator.DB) error {
	_, err := db.ExecContext(ctx, m.sql[selectLock], hash.Hash(m.names.Schema))
	return err
}

func (m *Migrator) SelectUnlock(ctx context.Context, db migrator.DB) error {
	_, err := db.ExecContext(ctx, m.sql[selectUnlock], hash.Hash(m.names.Schema))
	return err
}

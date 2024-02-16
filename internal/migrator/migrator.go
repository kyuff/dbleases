package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TableNames struct {
	Schema string
	Prefix string
}

func New(db DB, schema Schema) *Migrator {
	return &Migrator{
		db:     db,
		schema: schema,
	}
}

type Migrator struct {
	db     DB
	schema Schema
}

// Schema performs queries on the RDBMS at hand.
// This setup is a mess, and should be cleaned up at some point.
type Schema interface {
	Migrations() map[string]string
	CreateSchema(ctx context.Context, db DB) error
	CreateMigrationTable(ctx context.Context, db DB) error
	SelectMaxMigration(ctx context.Context, db DB) *sql.Row
	InsertMigrationRow(ctx context.Context, db DB, version uint32, fileName, sha string) error
	SelectUnlock(ctx context.Context, db DB) error
	SelectLock(ctx context.Context, db DB) error
}

func (m *Migrator) Migrate(ctx context.Context) error {
	err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer m.unlock()

	err = m.createMigrationTable(ctx)
	if err != nil {
		return err
	}

	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return err
	}

	migrations, err := parseMigrations(m.schema.Migrations())
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if migration.version <= currentVersion {
			continue
		}

		err = m.migrate(ctx, migration)
		if err != nil {
			return err
		}

		err = m.setCurrentVersion(ctx, migration)
		if err != nil {
			return err
		}

		currentVersion = migration.version
	}

	return nil
}

// lock blocks until the context times out or the lock is acquired
func (m *Migrator) lock(ctx context.Context) error {
	err := m.schema.SelectLock(ctx, m.db)
	if err != nil {
		return fmt.Errorf("lock failed: %s", err)
	}

	return nil
}

// unlock frees up the lock
func (m *Migrator) unlock() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err := m.schema.SelectUnlock(ctx, m.db)
	if err != nil {
		slog.ErrorContext(ctx, "[dbleases] Failed to unlock migration table")
	}
}

// migrationTableMutex guards against creating the _migration tables in parallel
var migrationTableMutex sync.Mutex

func (m *Migrator) createMigrationTable(ctx context.Context) error {
	migrationTableMutex.Lock()
	defer migrationTableMutex.Unlock()

	err := m.schema.CreateSchema(ctx, m.db)
	if err != nil {
		return err
	}

	err = m.schema.CreateMigrationTable(ctx, m.db)
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) setCurrentVersion(ctx context.Context, migration migrationVersion) error {
	return m.schema.InsertMigrationRow(ctx, m.db, migration.version, migration.fileName, migration.SHA512())
}

// getCurrentVersion returns 0 for no migrations run or >0 with the version
func (m *Migrator) getCurrentVersion(ctx context.Context) (uint32, error) {
	row := m.schema.SelectMaxMigration(ctx, m.db)
	var version sql.NullInt32
	err := row.Scan(&version)
	if err != nil {
		return 0, err
	}

	if !version.Valid {
		return 0, nil
	}

	return uint32(version.Int32), nil
}

func (m *Migrator) migrate(ctx context.Context, migration migrationVersion) error {
	_, err := m.db.ExecContext(ctx, migration.ddl)
	if err != nil {
		return err
	}
	return nil
}

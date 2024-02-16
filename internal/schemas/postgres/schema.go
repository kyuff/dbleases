package postgres

import (
	"embed"
	"errors"
	"fmt"

	"github.com/kyuff/dbleases/internal/tmpl"
)

const (
	createSchema         = "create_schema.tmpl"
	createMigrationTable = "create_schema_migrations.tmpl"
	insertMigrationRow   = "insert_migration_row.tmpl"
	selectMaxMigration   = "select_max_migration.tmpl"
	selectLock           = "select_lock.tmpl"
	selectUnlock         = "select_unlock.tmpl"
)

var migratorFiles = []string{
	createSchema,
	createMigrationTable,
	insertMigrationRow,
	selectMaxMigration,
	selectLock,
	selectUnlock,
}

const (
	insertLease         = "insert_lease.tmpl"
	selectRefreshLeases = "select_refresh_leases.tmpl"
	updateLeaseStatus   = "update_lease_status.tmpl"
	deleteLeases        = "delete_leases.tmpl"
)

var clientFiles = []string{
	selectRefreshLeases,
	insertLease,
	updateLeaseStatus,
	deleteLeases,
}

type tableNames struct {
	Schema string
	Prefix string
}

func parseAndValidate(fs embed.FS, names tableNames, expectedFiles []string) (map[string]string, error) {
	sqlMap, err := tmpl.Parse(fs, names)
	if err != nil {
		return nil, err
	}

	for _, fileName := range expectedFiles {
		_, ok := sqlMap[fileName]
		if !ok {
			err = errors.Join(err, fmt.Errorf("missing sql file: %s", fileName))
		}
	}

	return sqlMap, err
}

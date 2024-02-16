package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/kyuff/dbleases/internal/lease"
	"github.com/kyuff/dbleases/internal/migrator"
	"github.com/kyuff/dbleases/internal/rfc8601"
	"github.com/kyuff/dbleases/internal/schemas/postgres/sql"
	"github.com/kyuff/dbleases/internal/tmpl"
)

func New(ctx context.Context, db DB, schema, prefix string) (*Repository, error) {
	var names = tableNames{Schema: schema, Prefix: prefix}
	migratorSQL, err := parseAndValidate(sql.Migrator, names, migratorFiles)
	if err != nil {
		return nil, err
	}

	migrations, err := tmpl.Parse(sql.Migrations, names)
	if err != nil {
		return nil, err
	}

	clientSQL, err := parseAndValidate(sql.Client, names, clientFiles)
	if err != nil {
		return nil, err
	}

	m := migrator.New(db, &Migrator{sql: migratorSQL, names: names, migrations: migrations})
	err = m.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:  db,
		sql: clientSQL,
	}, nil
}

type Repository struct {
	db  DB
	sql map[string]string
}

func (s *Repository) InsertLease(ctx context.Context, clientID string, leaseName string, value int, ttl time.Duration, status lease.Status) error {
	_, err := s.db.ExecContext(ctx, s.sql[insertLease],
		leaseName,
		clientID,
		rfc8601.Format(ttl),
		status,
		value,
	)

	return err
}
func (s *Repository) GetAndRefreshLeases(ctx context.Context, names []string, clientID string, ttl time.Duration) ([]lease.Info, error) {
	rows, err := s.db.QueryContext(ctx, s.sql[selectRefreshLeases], names, clientID, rfc8601.Format(ttl))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var leases []lease.Info
	for rows.Next() {
		var info lease.Info
		err = rows.Scan(
			&info.Name,
			&info.ClientID,
			&info.TTL,
			&info.Status,
			&info.Value,
		)
		if err != nil {
			return nil, err
		}
		leases = append(leases, info)
	}
	return leases, nil

}

func (s *Repository) SetLeaseStatus(ctx context.Context, clientID string, leaseName string, value int, status lease.Status) error {
	_, err := s.db.ExecContext(ctx, s.sql[updateLeaseStatus], clientID, leaseName, value, status)
	return err
}
func (s *Repository) DeleteLeases(ctx context.Context, clientID string) error {
	_, err := s.db.ExecContext(ctx, s.sql[deleteLeases], clientID)
	return err
}

func (s *Repository) debugPrint(script string, args ...any) {
	fmt.Printf("SQL:\n%sVALUES: %q\n", s.sql[script], args)
}

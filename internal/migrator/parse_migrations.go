package migrator

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"slices"
	"strconv"
)

type migrationVersion struct {
	version  uint32
	fileName string
	ddl      string
}

func (m migrationVersion) SHA256() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(m.ddl)))
}

func (m migrationVersion) SHA512() string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(m.ddl)))
}

func parseMigrations(input map[string]string) ([]migrationVersion, error) {
	var migrations []migrationVersion
	for fileName, ddl := range input {
		version, err := parseFileVersion(fileName)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migrationVersion{
			version:  version,
			fileName: fileName,
			ddl:      ddl,
		})
	}

	slices.SortFunc(migrations, func(a, b migrationVersion) int {
		if a.version < b.version {
			return -1
		}
		if a.version > b.version {
			return 1
		}

		return 0
	})

	for i := 0; i < len(migrations); i++ {
		if uint32(i+1) != migrations[i].version {
			return nil, fmt.Errorf("migration %d not numbered in sequence: %s", i, migrations[i].fileName)
		}
	}

	return migrations, nil
}

func parseFileVersion(name string) (uint32, error) {
	if len(name) < 4 {
		return 0, fmt.Errorf("migration file name too short: %s", name)
	}

	n, err := strconv.Atoi(name[0:3])
	if err != nil {
		return 0, fmt.Errorf("file name must start with numbers %q: %s", name, err)
	}

	return uint32(n), nil
}

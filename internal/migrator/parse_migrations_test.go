package migrator

import (
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
)

func TestParseMigrations(t *testing.T) {

	t.Run("parse correct structure", func(t *testing.T) {
		// arrange
		var (
			input = map[string]string{
				"001_table_a.tmpl": "001_table_a.sql",
				"002_table_b.tmpl": "002_table_b.sql",
			}
		)

		// act
		got, err := parseMigrations(input)

		// assert
		assert.NoError(t, err)
		if assert.Equal(t, 2, len(got)) {
			assert.Equal(t, uint32(1), got[0].version)
			assert.Equal(t, "001_table_a.tmpl", got[0].fileName)
			assert.Match(t, ".*table_a.*", got[0].ddl)

			assert.Equal(t, uint32(2), got[1].version)
			assert.Equal(t, "002_table_b.tmpl", got[1].fileName)
			assert.Match(t, ".*table_b.*", got[1].ddl)
		}
	})

	t.Run("fail on missing version number", func(t *testing.T) {
		// arrange
		var (
			input = map[string]string{
				"001_table_a.tmpl": "001_table_a.sql",
				"002_table_b.tmpl": "002_table_b.sql",
				"004_table_b.tmpl": "004_table_b.sql",
			}
		)

		// act
		_, err := parseMigrations(input)

		// assert
		assert.Error(t, err)
	})

	t.Run("fail on wrong file name", func(t *testing.T) {
		// arrange
		var (
			input = map[string]string{
				"x_table_a.tmpl": "x_table_a.sql",
				"y_table_b.tmpl": "y_table_b.sql",
				"z_table_b.tmpl": "z_table_b.sql",
			}
		)

		// act
		_, err := parseMigrations(input)

		// assert
		assert.Error(t, err)
	})
}

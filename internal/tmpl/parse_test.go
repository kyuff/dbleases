package tmpl_test

import (
	"embed"
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
	"github.com/kyuff/dbleases/internal/tmpl"
)

//go:embed testdata/*.tmpl
var testdata embed.FS

func TestParse(t *testing.T) {
	// arrange
	type Data struct {
		Name string
	}
	var expected = map[string]string{
		"file_a.tmpl": "File A\nName: Hello world!\n",
		"file_x.tmpl": "File Hello world!\n",
	}

	// act
	got, err := tmpl.Parse(testdata, Data{Name: "Hello world!"})

	// assert
	assert.NoError(t, err)
	assert.EqualMap(t, expected, got)
}

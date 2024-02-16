package hash_test

import (
	"fmt"
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
	"github.com/kyuff/dbleases/internal/hash"
)

func TestHash(t *testing.T) {
	assert.Equal(t, 1423569895, hash.Hash("stream type"))
	assert.Equal(t, 389466507, hash.Hash("really long stream name that hopefully is not realistic"))
	assert.Equal(t, 635986699, hash.Hash("x----------------y---------------z"))
}

func TestMod(t *testing.T) {
	assert.Equal(t, 5, hash.Mod("mods to 5/20 (ab)", 20))
	assert.Equal(t, 10, hash.Mod("mods to 10/20 (ag)", 20))
	assert.Equal(t, 15, hash.Mod("mods to 15/20 (dt)", 20))

	assert.Equal(t, 50, hash.Mod("mods to 50/100 (bc)", 100))

	t.Logf("Hash: %q", findCorrectMod(15, 100))
}

func findCorrectMod(val, max int) []string {
	const (
		template = "%s hash %d/%d"
		seed     = "abcdefghijklmnopqrstuvwxyz"
	)

	var matches []string
	for i := 0; i < len(seed); i++ {
		for j := 0; j < len(seed); j++ {
			attempt := fmt.Sprintf(template, string(seed[i])+string(seed[j]), val, max)
			if hash.Mod(attempt, max) == val {
				matches = append(matches, attempt)
			}
		}
	}

	return matches
}

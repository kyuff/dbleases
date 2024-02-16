package dbleases

import (
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
)

func TestPresentIntegers(t *testing.T) {
	var assertPresentation = func(t *testing.T, expect string, input []int) {
		t.Helper()
		got := presentIntegers(input)
		if !assert.Equal(t, expect, got) {
			t.Logf("%v", input)
		}
	}
	t.Run("single item", func(t *testing.T) {
		assertPresentation(t, "1/5", []int{5})
	})

	t.Run("single connected list", func(t *testing.T) {
		assertPresentation(t, "3/5-7", []int{5, 6, 7})
	})

	t.Run("two connected lists", func(t *testing.T) {
		assertPresentation(t, "6/5-7,10-12", []int{5, 6, 7, 10, 11, 12})
	})

	t.Run("single item between two connected", func(t *testing.T) {
		assertPresentation(t, "6/5-6,8,10-12", []int{5, 6, 8, 10, 11, 12})
	})

	t.Run("single item at start", func(t *testing.T) {
		assertPresentation(t, "4/8,10-12", []int{8, 10, 11, 12})
	})

	t.Run("single item at end", func(t *testing.T) {
		assertPresentation(t, "5/7-10,12", []int{7, 8, 9, 10, 12})
	})

	t.Run("multiple single items", func(t *testing.T) {
		assertPresentation(t, "3/1,5,12", []int{1, 5, 12})
	})

	t.Run("two small ranges", func(t *testing.T) {
		assertPresentation(t, "4/1-2,4-5", []int{1, 2, 4, 5})
	})

	t.Run("empty list", func(t *testing.T) {
		assertPresentation(t, "0/-", nil)
	})
}

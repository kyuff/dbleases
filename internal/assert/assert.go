package assert

import (
	"reflect"
	"regexp"
	"testing"
	"time"
)

func EqualMap[K, V comparable, T ~map[K]V](t *testing.T, expected, got T) bool {
	t.Helper()
	if len(expected) != len(got) {
		t.Logf(`
Size difference
Expected: %d
     Got: %d`, len(expected), len(got))
		t.Fail()
		return false
	}

	for k, v := range expected {
		gotValue, ok := got[k]
		if !ok {
			t.Logf(`
Expected key %v not found`, k)
			t.Fail()
			continue
		}

		if v != gotValue {
			t.Logf(`
Expected value for key key %v not equal
Expected: %v
     Got: %v`, k, v, gotValue)
			t.Fail()
		}
	}

	return t.Failed()
}

func Equal[T comparable](t *testing.T, expected, got T) bool {
	t.Helper()
	if expected != got {
		t.Logf(`
Items was not equal
Expected: %v
     Got: %v`, expected, got)
		t.Fail()
		return false
	}
	return true
}

func EqualSliceWithin[V comparable, T []V](t *testing.T, within time.Duration, expected T, fn func() T) bool {
	t.Helper()
	iterations := 100
	ticker := time.NewTicker(within / time.Duration(iterations))
	defer ticker.Stop()

	var got T
	for range ticker.C {
		iterations = iterations - 1
		if iterations <= 0 {
			break
		}
		got = fn()
		if len(expected) != len(got) {
			continue
		}

		match := true
		for i := 0; i < len(expected); i++ {
			if expected[i] != got[i] {
				match = false
			}
		}
		if !match {
			continue
		}

		return true
	}

	return EqualSlice(t, expected, got)
}

func EqualSlice[V comparable, T []V](t *testing.T, expected, got T) bool {
	t.Helper()
	match := true
	if len(expected) != len(got) {
		match = false
	} else {
		for i := 0; i < len(expected); i++ {
			if expected[i] != got[i] {
				match = false
			}
		}
	}

	if match {
		return true
	}

	t.Logf(`
Slices was not equal
Expected: %v
     Got: %v`, expected, got)
	t.Fail()
	return false
}

func NotEqual[T comparable](t *testing.T, unexpected, got T) bool {
	t.Helper()
	if unexpected == got {
		t.Logf(`
Items was equal
Expected: %v
     Got: %v`, unexpected, got)
		t.Fail()
		return false
	}
	return true
}

func NotNil(t *testing.T, got any) bool {
	t.Helper()
	if reflect.ValueOf(got).IsNil() {
		t.Logf("Expected a value, but got nil")
		t.Fail()
		return false
	}

	return true
}

func Nil(t *testing.T, got any) bool {
	t.Helper()
	if !reflect.ValueOf(got).IsNil() {
		t.Logf("Expected nil, but got a value: %#v", got)
		t.Fail()
		return false
	}

	return true
}
func Match[T ~string](t *testing.T, expectedRE string, got T) bool {
	t.Helper()
	re, err := regexp.Compile(expectedRE)
	if err != nil {
		t.Fatalf("unexpected regexp: %s", err)
		return false
	}

	match := re.MatchString(string(got))
	if !match {
		t.Logf(`
Must match %q
       Got %q`, expectedRE, got)
		t.Fail()
		return false
	}

	return true
}

func OneOf[T comparable](t *testing.T, items []T, got T) bool {
	t.Helper()
	var found = false
	for _, item := range items {
		if item == got {
			found = true
		}
	}

	if !found {
		t.Logf("Input list: %v", items)
		t.Logf("Did not contain item: %v", got)
		t.Fail()
		return false
	}

	return true
}

func NoneZero[T any, E ~[]T](t *testing.T, got E) bool {
	t.Helper()
	for _, e := range got {
		if reflect.ValueOf(e).IsZero() {
			return false
		}
	}

	return true
}

func TimeWithinWindow(t *testing.T, expected time.Time, got time.Time, window time.Duration) bool {
	var (
		from = expected.Add(-1 * window)
		to   = expected.Add(window)
	)

	if got.Before(from) {
		t.Logf("Time was before the window by %s", from.Sub(got))
		t.Fail()
	}

	if got.After(to) {
		t.Logf("Time was after the window by %s", got.Sub(to))
		t.Fail()
	}

	return true
}

func NoError(t *testing.T, got error) bool {
	t.Helper()
	if got != nil {
		t.Logf("Unexpected error: %s", got)
		t.Fail()
		return false
	}

	return true
}

func Error(t *testing.T, got error) bool {
	t.Helper()
	if got == nil {
		t.Logf("Expected error: %s", got)
		t.Fail()
		return false
	}

	return true
}

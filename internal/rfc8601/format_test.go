package rfc8601_test

import (
	"testing"
	"time"

	"github.com/kyuff/dbleases/internal/assert"
	"github.com/kyuff/dbleases/internal/rfc8601"
)

func TestFormat(t *testing.T) {
	var tests = map[time.Duration]string{
		time.Second + 600*time.Millisecond:                         "PT1S",
		time.Second:                                                "PT1S",
		time.Second + 2*time.Minute:                                "PT2M1S",
		time.Second + 2*time.Minute + 3*time.Hour:                  "PT3H2M1S",
		time.Second + 2*time.Minute + 3*time.Hour + 4*24*time.Hour: "PT99H2M1S",
		time.Second + 2*time.Minute + 3*time.Hour + 400*time.Hour:  "PT403H2M1S",
	}

	for input, expected := range tests {
		t.Run(input.String(), func(t *testing.T) {
			assert.Equal(t, expected, rfc8601.Format(input))
		})
	}
}

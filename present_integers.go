package dbleases

import (
	"bytes"
	"sort"
	"strconv"
)

func presentIntegers(list []int) string {
	const (
		hyphen = '-'
		comma  = ','
	)
	sort.Ints(list)

	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(len(list)))
	buf.WriteRune('/')

	switch len(list) {
	case 0:
		buf.WriteRune(hyphen)
		return buf.String()
	case 1:
		buf.WriteString(strconv.Itoa(list[0]))
		return buf.String()
	}
	var lastIndex = len(list) - 1
	var isRangeStart = false
	var prev = list[0]
	for i := 1; i < len(list); i++ {
		var current = list[i]

		if prev+1 == current {
			if !isRangeStart {
				buf.WriteString(strconv.Itoa(prev))
				buf.WriteRune(hyphen)
			}

			if lastIndex == i {
				buf.WriteString(strconv.Itoa(current))
			}

			isRangeStart = true
		} else {
			buf.WriteString(strconv.Itoa(prev))
			buf.WriteRune(comma)
			isRangeStart = false
			if lastIndex == i {
				buf.WriteString(strconv.Itoa(current))
			}
		}
		prev = current
	}

	return buf.String()
}

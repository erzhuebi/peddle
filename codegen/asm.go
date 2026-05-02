package codegen

import (
	"strconv"
	"strings"
)

func asmStringBytes(s string) string {
	var parts []string

	for _, r := range s {
		switch r {
		case '\n':
			parts = append(parts, "13")
		case '"':
			parts = append(parts, "34")
		default:
			parts = append(parts, strconv.Quote(string(r)))
		}
	}

	return strings.Join(parts, ",")
}

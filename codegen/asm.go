package codegen

import (
	"fmt"
)

func asmStringBytes(s string) string {
	out := ""

	for i := 0; i < len(s); i++ {
		c := s[i]

		// handle newline escape
		if c == '\n' {
			out += "13"
		} else {
			// normalize to uppercase (C64 default charset)
			if c >= 'a' && c <= 'z' {
				c = c - 'a' + 'A'
			}

			out += fmt.Sprintf("%d", c)
		}

		if i != len(s)-1 {
			out += ","
		}
	}

	return out
}

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

func (g *Generator) emitLongBranch(op string, target string) error {
	inverse, err := inverseBranchOp(op)
	if err != nil {
		return err
	}

	skip := g.newLabel()

	g.emit(fmt.Sprintf("    %s %s", inverse, skip))
	g.emit(fmt.Sprintf("    jmp %s", target))
	g.emit(skip + ":")

	return nil
}

func inverseBranchOp(op string) (string, error) {
	switch op {
	case "beq":
		return "bne", nil
	case "bne":
		return "beq", nil
	case "bcc":
		return "bcs", nil
	case "bcs":
		return "bcc", nil
	case "bmi":
		return "bpl", nil
	case "bpl":
		return "bmi", nil
	default:
		return "", fmt.Errorf("unsupported branch op %q", op)
	}
}

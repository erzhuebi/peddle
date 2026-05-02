package codegen

import "fmt"

func (g *Generator) addLiteral(value string) string {
	label := fmt.Sprintf("literal_%d", len(g.literals))
	g.literals = append(g.literals, value)
	return label
}

func (g *Generator) newLabel() string {
	g.labelCounter++
	return fmt.Sprintf("L%d", g.labelCounter)
}

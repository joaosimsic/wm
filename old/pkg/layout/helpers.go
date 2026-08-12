package layout

import (
	"fmt"
	"strings"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ContainerDebug(c *Container, indent string) string {
	if c == nil {
		return indent + "nil\n"
	}
	switch c.Type {
	case ContainerLeaf:
		title := "nil"
		if c.Window != nil {
			title = c.Window.Title
		}
		return fmt.Sprintf("%sLeaf(%s)\n", indent, title)
	case ContainerVSplit:
		var b strings.Builder
		fmt.Fprintf(&b, "%sVSplit[\n", indent)
		for _, ch := range c.Children {
			b.WriteString(ContainerDebug(ch, indent+"  "))
		}
		fmt.Fprintf(&b, "%s]\n", indent)
		return b.String()
	case ContainerHSplit:
		var b strings.Builder
		fmt.Fprintf(&b, "%sHSplit[\n", indent)
		for _, ch := range c.Children {
			b.WriteString(ContainerDebug(ch, indent+"  "))
		}
		fmt.Fprintf(&b, "%s]\n", indent)
		return b.String()
	}
	return indent + "???\n"
}

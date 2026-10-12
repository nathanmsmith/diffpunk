package private

import (
	"strings"
)

func writePatchLine(out *strings.Builder, line patchLine) {
	out.WriteByte(line.kind)
	out.WriteString(line.text)
	if !strings.HasSuffix(line.text, "\n") {
		out.WriteString("\n\\ No newline at end of file\n")
	}
}

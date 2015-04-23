package dot

import (
	"bytes"
	"fmt"
)

type GraphWriter struct {
	bytes.Buffer
	indent    int
	indentStr string
}

func NewGraphWriter() *GraphWriter {
	w := &GraphWriter{
		indent: 0,
	}
	w.init()
	return w
}

func (w *GraphWriter) Printf(s string, args ...interface{}) {
	w.WriteString(w.indentStr + fmt.Sprintf(s, args...))
}

func (w *GraphWriter) Indent() {
	w.indent++
	w.init()
}

func (w *GraphWriter) Unindent() {
	w.indent--
	w.init()
}

func (w *GraphWriter) Newline() {
	w.WriteString("\n")
}

func (w *GraphWriter) init() {
	indentBuf := new(bytes.Buffer)
	for i := 0; i < w.indent; i++ {
		indentBuf.WriteString("\t")
	}
	w.indentStr = indentBuf.String()
}

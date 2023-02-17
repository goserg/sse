package sse

import (
	"bytes"
	"io"
	"strconv"
)

type message struct {
	id    int
	event string
	data  []byte
}

func (m *message) writeTo(w io.Writer) error {
	var buf bytes.Buffer

	if m.id != 0 {
		buf.WriteString("id")
		buf.WriteString(strconv.Itoa(m.id))
		buf.WriteString("\n")
	}
	if m.event != "" {
		buf.WriteString("event")
		buf.WriteString(m.event)
		buf.WriteString("\n")
	}
	if len(m.data) != 0 {
		buf.WriteString("data")
		buf.Write(m.data)
		buf.WriteString("\n")
	}

	_, err := buf.WriteTo(w)
	return err
}

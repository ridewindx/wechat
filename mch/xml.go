package mch

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
)

// DecodeXML decodes xml from io.Reader and returns the first-level sub-node key-value set.
// If the first-level sub-node contains child nodes, skip it.
func DecodeXML(r io.Reader) (m map[string]string, err error) {
	m = make(map[string]string)
	var (
		decoder = xml.NewDecoder(r)
		depth   = 0
		token   xml.Token
		key     string
		value   bytes.Buffer
	)
	for {
		token, err = decoder.Token()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		switch v := token.(type) {
		case xml.StartElement:
			depth++
			switch depth {
			case 2:
				key = v.Name.Local
				value.Reset()
			case 3:
				if err = decoder.Skip(); err != nil {
					return
				}
				depth--
				key = "" // key == "" indicates that the node with depth==2 has children
			}
		case xml.CharData:
			if depth == 2 && key != "" {
				value.Write(v)
			}
		case xml.EndElement:
			if depth == 2 && key != "" {
				m[key] = value.String()
			}
			depth--
		}
	}
}

const rootName = "xml"

type writer interface {
	io.Writer
	io.ByteWriter
	WriteString(s string) (n int, err error)
}

// EncodeXML encodes map[string]string to io.Writer with xml format.
func EncodeXML(w io.Writer, m map[string]string) (err error) {
	var b writer
	switch v := w.(type) {
	case writer:
		b = v
	default:
		b = bufio.NewWriterSize(w, 256)
	}

	if err = b.WriteByte('<'); err != nil {
		return
	}
	if _, err = b.WriteString(rootName); err != nil {
		return
	}
	if err = b.WriteByte('>'); err != nil {
		return
	}

	for k, v := range m {
		if err = b.WriteByte('<'); err != nil {
			return
		}
		if _, err = b.WriteString(k); err != nil {
			return
		}
		if err = b.WriteByte('>'); err != nil {
			return
		}

		if err = xml.EscapeText(b, []byte(v)); err != nil {
			return
		}

		if _, err = b.WriteString("</"); err != nil {
			return
		}
		if _, err = b.WriteString(k); err != nil {
			return
		}
		if err = b.WriteByte('>'); err != nil {
			return
		}
	}

	if _, err = b.WriteString("</"); err != nil {
		return
	}
	if _, err = b.WriteString(rootName); err != nil {
		return
	}
	if err = b.WriteByte('>'); err != nil {
		return
	}

	if v, ok := b.(*bufio.Writer); ok {
		return v.Flush()
	}

	return
}

package mail

import (
	"bytes"
	"encoding/base64"
	"mime/quotedprintable"
	"strings"
)

// base64Encode base64 encodes the provided text with line wrapping
func base64Encode(text []byte) []byte {
	// create buffer
	buf := new(bytes.Buffer)

	// create base64 encoder that linewraps
	encoder := base64.NewEncoder(base64.StdEncoding, &base64LineWrap{writer: buf})

	// write the encoded text to buf
	encoder.Write(text)
	encoder.Close()

	return buf.Bytes()
}

// qpEncode uses the quoted-printable encoding to encode the provided text
func qpEncode(text []byte) []byte {
	// create buffer
	buf := new(bytes.Buffer)

	encoder := quotedprintable.NewWriter(buf)

	encoder.Write(text)
	encoder.Close()

	return buf.Bytes()
}

func encodeHeader(text string, charset string, usedChars int) string {
	// create buffer
	buf := new(bytes.Buffer)

	// encode
	encoder := newEncoder(buf, charset, usedChars)
	encoder.encode([]byte(text))

	return buf.String()
}

func escapeQuotes(s string) string {
	quoteEscaper := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
	return quoteEscaper.Replace(s)
}

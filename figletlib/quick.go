package figletlib

import (
	"bytes"
)

const FONT_DIR = "figletlib/fonts"

func TXTToAscii(input string) string {
	// Create a byte slice
	var buf []byte

	// Create a bytes.Buffer, which implements io.Writer
	w := bytes.NewBuffer(buf)

	f, err := GetFontByName(FONT_DIR, "standard")
	if err != nil {
		return ""
	}

	FPrintMsg(w, input, f, 80, f.Settings(), "left")
	return w.String()
}

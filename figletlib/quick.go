package figletlib

import (
	"bytes"
)

const FONT_DIR = "figletlib/fonts"

func TXTToAscii(input, font, just string, cols int) (string, error) {
	// Create a byte slice
	var buf []byte

	// Create a bytes.Buffer, which implements io.Writer
	w := bytes.NewBuffer(buf)

	if font == "" {
		font = "standard"
	}
	if just == "" {
		just = "left"
	}

	f, err := GetFontByName(FONT_DIR, font)
	if err != nil {
		return "", err
	}
	if cols == 0 {
		cols = 80
	}

	FPrintMsg(w, input, f, cols, f.Settings(), just)
	return w.String(), nil
}

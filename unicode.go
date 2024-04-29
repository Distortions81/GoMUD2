package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/muesli/reflow/wrap"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const charsetSend = ";UTF-8;ISO88591;WINDOWS1252;LATIN1;MCP437;CP437;IBM437;MCP850;MCP858;MACROMAN;MACINTOSH;ASCII"

// use all caps!
var charsetList map[string]*charmap.Charmap = map[string]*charmap.Charmap{
	"ASCII":  charmap.ISO8859_1,
	"LATIN1": charmap.ISO8859_1,

	"ISO88591":  charmap.ISO8859_1,
	"ISO88592":  charmap.ISO8859_2,
	"ISO88593":  charmap.ISO8859_3,
	"ISO88594":  charmap.ISO8859_4,
	"ISO88595":  charmap.ISO8859_5,
	"ISO88596":  charmap.ISO8859_6,
	"ISO88597":  charmap.ISO8859_7,
	"ISO88598":  charmap.ISO8859_8,
	"ISO88599":  charmap.ISO8859_9,
	"ISO885910": charmap.ISO8859_10,
	"ISO885913": charmap.ISO8859_13,
	"ISO885914": charmap.ISO8859_14,
	"ISO885915": charmap.ISO8859_15,
	"ISO885916": charmap.ISO8859_16,

	"MACROMAN":  charmap.Macintosh,
	"MACINTOSH": charmap.Macintosh,

	"MCP037": charmap.CodePage037,

	"MCP437":           charmap.CodePage437,
	"IBM437":           charmap.CodePage437,
	"437":              charmap.CodePage437,
	"CP437":            charmap.CodePage437,
	"CSPC8CODEPAGE437": charmap.CodePage437,

	"MCP850": charmap.CodePage850,
	"MCP852": charmap.CodePage852,
	"MCP855": charmap.CodePage855,
	"MCP858": charmap.CodePage858,
	"MCP860": charmap.CodePage860,
	"MCP862": charmap.CodePage862,
	"MCP863": charmap.CodePage863,
	"MCP865": charmap.CodePage865,
	"MCP866": charmap.CodePage866,

	"MCP1047": charmap.CodePage1047,
	"MCP1140": charmap.CodePage1140,

	"WINDOWS874":  charmap.Windows874,
	"WINDOWS1250": charmap.Windows1250,
	"WINDOWS1251": charmap.Windows1251,
	"WINDOWS1252": charmap.Windows1252,
	"WINDOWS1253": charmap.Windows1253,
	"WINDOWS1254": charmap.Windows1254,
	"WINDOWS1255": charmap.Windows1255,
	"WINDOWS1256": charmap.Windows1256,
	"WINDOWS1257": charmap.Windows1257,
	"WINDOWS1258": charmap.Windows1258,
}

func encodeFromUTF(cmap *charmap.Charmap, input string) []byte {
	f := wrap.NewWriter(0)
	f.PreserveSpace = true
	f.Newline = []rune{'\n'}
	f.KeepNewlines = true
	f.Write([]byte(input))
	var tmp io.Reader = strings.NewReader(f.String())
	tmp = transform.NewReader(tmp, cmap.NewEncoder()) // encode bytes to cmap
	encBytes, _ := io.ReadAll(tmp)
	return encBytes
}

func encodeToUTF(cmap *charmap.Charmap, input []byte) string {
	var d io.Reader = bytes.NewReader(input) // each line as string
	e := cmap.NewDecoder()
	utf8 := transform.NewReader(d, e) // decode from cmap to UTF-8
	decBytes, _ := io.ReadAll(utf8)
	return string(decBytes)
}

func setCharset(desc *descData) {
	if strings.EqualFold(desc.telnet.charset, "UTF8") {
		desc.telnet.utf = true
	} else {

		//Check if we have the charset
		charSetSearch := charsetList[desc.telnet.charset]
		if charSetSearch != nil {
			desc.telnet.charMap = charSetSearch
		} else {
			//Otherwise look for partial match
			for str, cmap := range charsetList {
				if strings.HasSuffix(desc.telnet.charset, str) {
					desc.telnet.charMap = cmap
					break
				}
			}
		}
	}
}

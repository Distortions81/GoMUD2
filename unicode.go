package main

import (
	"strings"

	"golang.org/x/text/encoding/charmap"
)

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

func convertText(charmap *charmap.Charmap, data []byte) []byte {
	var tmp []byte
	for _, myRune := range data {

		enc := charmap.NewEncoder()
		win, err := enc.String(string(myRune))
		if err != nil {
			tmp = append(tmp, []byte("?")...)
			continue
		}
		tmp = append(tmp, []byte(win)...)
	}
	return tmp
}

func convertByte(charmap *charmap.Charmap, data byte) byte {
	enc := charmap.NewEncoder()
	win, err := enc.String(string(data))
	if err != nil {
		win = "?"
	}
	bytes := []byte(win)
	return bytes[0]
}

func setCharset(desc *descData) {
	if strings.EqualFold(desc.telnet.charset, "UTF-8") {
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

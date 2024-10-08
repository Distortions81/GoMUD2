package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const ANSI_ESC = "\033["
const ANSI_RESET = ANSI_ESC + "0m"
const NEWLINE = "\r\n"

func xcolorHelp() {
	var outbuf string
	var lineBuf []string
	outbuf = outbuf + fmt.Sprintf("{x%-36v %v{k"+NEWLINE, "Extended colors:", "Pastel colors:")
	for _, line := range colorSwatch {
		buf := ""
		for _, color := range line {
			buf = buf + fmt.Sprintf("{%03v%03v ", color+300, color)
		}
		lineBuf = append(lineBuf, buf)
	}

	for i, line := range colorPastelSwatch {
		buf := " "
		for _, color := range line {
			buf = buf + fmt.Sprintf("{%03v%03v ", color+300, color)
		}
		lineBuf[i] = lineBuf[i] + buf
	}
	outbuf = outbuf + strings.Join(lineBuf, NEWLINE)
	outbuf = outbuf + NEWLINE
	outbuf = outbuf + "{xGrayscale:{k" + NEWLINE
	for _, line := range graySwatch {
		buf := " "
		for _, color := range line {
			buf = buf + fmt.Sprintf("{%03v%03v ", color+300, color)
		}
		buf = buf + NEWLINE
		outbuf = outbuf + buf
	}
	outbuf = outbuf + NEWLINE
	outbuf = outbuf + "{xSyntax: {{088{088Hello{x{{x. Requires 3 digits, padded with 0s if needed."
	outbuf = outbuf + NEWLINE
	outbuf = outbuf + "Background colors: add 300 to the number: {{388{388Hello{x{{x."

	for _, file := range helpFiles {
		if strings.EqualFold(file.Topic, "basics") {

			newHelp := helpData{Name: "xcolor", Keywords: []string{"ANSI", "color", "extended", "256"}, Authors: []string{"System"}, Text: outbuf, topic: file}

			//Update if found, otherwise create
			for h, help := range file.Helps {
				if strings.EqualFold(help.Name, "xcolor") {
					newHelp.Created = time.Now().UTC()
					newHelp.Modified = time.Now().UTC()
					file.Helps[h] = newHelp
					return
				}
			}
			file.Helps = append(file.Helps, newHelp)
			file.dirty = true

		}
	}

}

var colorTable map[byte]*ctData = map[byte]*ctData{
	//bg colors
	'!': {code: "40", isBG: true}, //black
	'@': {code: "41", isBG: true}, //red
	'#': {code: "42", isBG: true}, //green
	'$': {code: "43", isBG: true}, //yellow
	'%': {code: "44", isBG: true}, //blue
	'^': {code: "45", isBG: true}, //magenta
	'&': {code: "46", isBG: true}, //cyan
	'*': {code: "47", isBG: true}, //white

	'k': {code: "30", isFG: true, removeBold: true}, //black
	'r': {code: "31", isFG: true, removeBold: true}, //red
	'g': {code: "32", isFG: true, removeBold: true}, //green
	'y': {code: "33", isFG: true, removeBold: true}, //yellow
	'b': {code: "34", isFG: true, removeBold: true}, //blue
	'm': {code: "35", isFG: true, removeBold: true}, //magenta
	'c': {code: "36", isFG: true, removeBold: true}, //cyan
	'w': {code: "37", isFG: true, removeBold: true}, //white

	'K': {code: "30", isFG: true, style: bold}, //black
	'R': {code: "31", isFG: true, style: bold}, //red
	'G': {code: "32", isFG: true, style: bold}, //green
	'Y': {code: "33", isFG: true, style: bold}, //yellow
	'B': {code: "34", isFG: true, style: bold}, //blue
	'M': {code: "35", isFG: true, style: bold}, //magenta
	'C': {code: "36", isFG: true, style: bold}, //cyan
	'W': {code: "37", isFG: true, style: bold}, //white

	'+': {code: "1", disCode: "22", isStyle: true, style: bold},
	'/': {code: "3", disCode: "23", isStyle: true, style: italic},
	'_': {code: "4", disCode: "24", isStyle: true, style: underline},
	'=': {code: "7", disCode: "27", isStyle: true, style: inverse},
	'-': {code: "9", disCode: "29", isStyle: true, style: strike},
}

func ColorRemove(in []byte) []byte {

	//Process {n newlines first
	var out []byte
	inLen := len(in)
	for x := 0; x < inLen; x++ {
		if in[x] == '{' {
			x++
			if x < inLen {
				if in[x] == 'n' {
					out = append(out, []byte(NEWLINE)...)
				} else if in[x] == '{' {
					out = append(out, '{')
				}
				continue
			}
		} else {
			out = append(out, in[x])
		}
	}
	return out
}

const (
	COLOR_NONE = iota
	COLOR_16
	COLOR_256
	COLOR_TRUE
)

// Combines multiple color codes, allows styles to be toggled on and off and ignores any code that would set/unset a state that is already set/unset
func ANSIColor(in []byte, colorMode int) []byte {
	var s ansiState

	//Process {n newlines first
	var out []byte
	inLen := len(in)
	for x := 0; x < inLen; x++ {
		if in[x] == '{' {
			x++
			if x < inLen {
				// escaped {{
				if in[x] == 'n' {
					out = append(out, []byte(NEWLINE)...)
					continue
				} else {
					out = append(out, '{')
					out = append(out, in[x])
				}
			}
		} else {
			out = append(out, in[x])
		}
	}
	in = out

	//Process ANSI 256 color
	out = []byte{}
	inLen = len(in)
	var ext []byte
	for x := 0; x < inLen; x++ {

		//Escape code
		if in[x] == '{' {
			x++
			if x < inLen {
				//Escaped {
				if in[x] == '{' {
					out = append(out, '{')
					continue
					//Color reset
				} else if in[x] == 'x' {
					if s.hasColor {
						s.resetState()

						out = append(out, []byte(ANSI_RESET)...)
					}
					continue
				}

				var cVal *ctData
				//extended color
				if in[x] >= '0' && in[x] <= '9' {
					ext = []byte{in[x]}
					x++
					if in[x] >= '0' && in[x] <= '9' {
						ext = append(ext, in[x])
						x++
						if in[x] >= '0' && in[x] <= '9' {
							ext = append(ext, in[x])
						}

					}
					cnum, _ := strconv.ParseInt(string(ext), 10, 64)
					if cnum > 300 {
						cVal = &ctData{isFG: true, extended: true, code: "48;5;" + strconv.FormatInt(cnum-300, 10)}
						if colorMode < COLOR_256 {
							rcode := color256to16[int(cnum-300)]
							cVal = colorTable[rcode]
						}
					} else {
						cVal = &ctData{isFG: true, extended: true, code: "38;5;" + strconv.FormatInt(cnum, 10)}
						if colorMode < COLOR_256 {
							rcode := color256to16[int(cnum)]
							cVal = colorTable[rcode]
						}
					}
				} else {
					//Look up color/style
					cVal = colorTable[in[x]]
				}

				//If redundant, skip
				if s.lastVal == cVal {
					continue
				}
				s.lastVal = cVal

				//Not a valid code, skips
				if cVal == nil {
					continue
				}

				//If new FG color, set it.
				if cVal.isFG && s.curColor != cVal.code {
					s.nextColor = cVal.code
				}
				//If new BG color, set it.
				if cVal.isBG && s.curBGColor != cVal.code {
					s.nextBGColor = cVal.code
				}

				//Toggle styles such as italic
				if !cVal.isFG {
					s.nextStyle.toggleFlag(cVal.style)
				} else {
					//Otherwise, if bold FG color, add add flag only.
					s.nextStyle.addFlag(cVal.style)
				}
				//If we are switching from bold FG color
				//to non-bold FG color, remove bold (do not toggle)
				if cVal.removeBold {
					s.nextStyle.clearFlag(bold)
				}
				continue
			} else {
				break
			}
		} else {
			s.lastVal = nil

			//If we have a new character and the color or style has changed...
			if s.nextColor != "" || s.nextBGColor != "" || s.nextStyle != s.curStyle {
				var cOut []byte

				//If destination style AND color is default/empty, use [m to save space
				if (s.nextStyle == 0) && (s.nextColor == "") && (s.nextBGColor == "") {
					out = append(out, []byte(ANSI_RESET)...)
					out = append(out, in[x])
					s.resetState()
					continue
				} else if s.nextStyle == 0 && s.curStyle != 0 {
					//If we had a style, but now we do not set style to 0
					cOut = append(cOut, "0"...)
				} else if s.nextStyle.hasFlag(bold) && !s.curStyle.hasFlag(bold) {
					//Add bold style
					cOut = append(cOut, colorTable['+'].code...)
				} else if !s.nextStyle.hasFlag(bold) && s.curStyle.hasFlag(bold) {
					//Remove bold style
					cOut = append(cOut, colorTable['+'].disCode...)
				} else if s.nextStyle.hasFlag(italic) && !s.curStyle.hasFlag(italic) {
					//Add italic style
					cOut = append(cOut, colorTable['/'].code...)
				} else if !s.nextStyle.hasFlag(italic) && s.curStyle.hasFlag(italic) {
					//Remove italic style
					cOut = append(cOut, colorTable['/'].disCode...)
				} else if s.nextStyle.hasFlag(underline) && !s.curStyle.hasFlag(underline) {
					//Add underline style
					cOut = append(cOut, colorTable['_'].code...)
				} else if !s.nextStyle.hasFlag(underline) && s.curStyle.hasFlag(underline) {
					//Remove underline style
					cOut = append(cOut, colorTable['_'].disCode...)
				} else if s.nextStyle.hasFlag(inverse) && !s.curStyle.hasFlag(inverse) {
					//Add inverse style
					cOut = append(cOut, colorTable['='].code...)
				} else if !s.nextStyle.hasFlag(inverse) && s.curStyle.hasFlag(inverse) {
					//Remove inverse style
					cOut = append(cOut, colorTable['='].disCode...)
				} else if s.nextStyle.hasFlag(strike) && !s.curStyle.hasFlag(strike) {
					//Add strike style
					cOut = append(cOut, colorTable['-'].code...)
				} else if !s.nextStyle.hasFlag(strike) && s.curStyle.hasFlag(strike) {
					//Remove strike style
					cOut = append(cOut, colorTable['-'].disCode...)
				}
				//Add BG color if state changed
				if s.nextBGColor != s.curBGColor {
					if len(cOut) > 0 {
						cOut = append(cOut, ';')
					}
					cOut = append(cOut, []byte(s.nextBGColor)...)
					s.hasColor = true
					s.nextBGColor = ""
				}
				//Add FG color if state changed
				if s.nextColor != s.curColor {
					if len(cOut) > 0 {
						cOut = append(cOut, ';')
					}
					cOut = append(cOut, []byte(s.nextColor)...)
					s.hasColor = true
					s.nextColor = ""
				}
				//If we have a color code, end ANSI sequence
				if len(cOut) > 0 {
					cOut = append(cOut, 'm')

					//Set current state from new state
					s.curStyle = s.nextStyle
					s.curColor = s.nextColor
					s.curBGColor = s.nextBGColor

					s.hasColor = true
					//escape code
					out = append(out, []byte(ANSI_ESC)...)
					//ansi code
					out = append(out, cOut...)
				}
			}
		}
		//Append text
		out = append(out, in[x])

	}

	if s.hasColor {
		//If our state has a non-default color or style, reset it at the end
		out = append(out, []byte(ANSI_RESET)...)
	}
	return out
}

type ansiState struct {
	curStyle, nextStyle Bitmask

	curColor,
	curBGColor,

	nextColor,
	nextBGColor string

	hasColor bool
	lastVal  *ctData
}

func (state *ansiState) resetState() {
	state.curStyle = 0
	state.nextStyle = 0

	state.curColor = ""
	state.nextColor = ""

	state.curBGColor = ""
	state.nextBGColor = ""

	state.hasColor = false
}

const (
	bold = iota
	italic
	underline
	inverse
	strike
)

type ctData struct {
	code, disCode string
	style         int

	isBG, isFG, removeBold,
	isStyle, extended bool
}

var colorSwatch [][]int = [][]int{
	{52, 88, 124, 160, 196, 203, 210, 217, 224},
	{52, 88, 124, 160, 202, 209, 216, 223, 230},
	{52, 88, 124, 166, 208, 215, 222, 229, 230},
	{52, 88, 130, 172, 214, 221, 228, 229, 230},
	{52, 94, 136, 178, 220, 227, 228, 229, 230},
	{58, 100, 142, 184, 226, 227, 228, 229, 231},
	{22, 64, 106, 148, 190, 227, 228, 229, 230},
	{22, 28, 70, 112, 154, 191, 228, 229, 230},
	{22, 28, 34, 76, 118, 155, 192, 229, 230},
	{22, 28, 34, 40, 82, 119, 156, 193, 230},
	{22, 28, 34, 40, 46, 83, 120, 157, 194},
	{22, 28, 34, 40, 47, 84, 121, 158, 195},
	{22, 28, 34, 41, 48, 85, 122, 159, 195},
	{22, 28, 35, 42, 49, 86, 123, 159, 195},
	{22, 29, 36, 43, 50, 87, 123, 159, 195},
	{23, 30, 37, 44, 51, 87, 123, 159, 195},
	{17, 24, 31, 38, 45, 87, 123, 159, 195},
	{17, 18, 25, 32, 39, 81, 123, 159, 195},
	{17, 18, 19, 26, 33, 75, 117, 159, 195},
	{17, 18, 19, 20, 27, 69, 111, 153, 195},
	{17, 18, 19, 20, 21, 63, 105, 147, 189},
	{17, 18, 19, 20, 57, 99, 141, 183, 225},
	{17, 18, 19, 56, 93, 135, 177, 219, 225},
	{17, 18, 55, 92, 129, 171, 213, 219, 225},
	{17, 54, 91, 128, 165, 207, 213, 219, 225},
	{53, 90, 127, 164, 201, 207, 213, 219, 225},
	{52, 89, 126, 163, 200, 207, 213, 219, 225},
	{52, 88, 125, 162, 199, 206, 213, 219, 225},
	{52, 88, 124, 161, 198, 205, 212, 219, 225},
	{52, 88, 124, 160, 197, 204, 211, 218, 225},
}

var colorPastelSwatch [][]int = [][]int{
	{95, 131, 167, 174, 181},
	{95, 131, 173, 180, 187},
	{95, 137, 179, 186, 187},
	{101, 143, 185, 186, 187},
	{65, 107, 149, 186, 187},
	{65, 71, 113, 150, 187},
	{65, 71, 77, 114, 151},
	{65, 71, 78, 115, 152},
	{65, 72, 79, 116, 152},
	{66, 73, 80, 116, 152},
	{60, 67, 74, 116, 152},
	{60, 61, 68, 110, 152},
	{60, 61, 62, 104, 146},
	{60, 61, 98, 140, 182},
	{60, 97, 134, 176, 182},
	{96, 133, 170, 176, 182},
	{95, 132, 169, 176, 182},
	{95, 131, 168, 175, 182},
}

var graySwatch [][]int = [][]int{
	{232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243},
	{244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
}

// uint8 to 16-color symbol byte
var color256to16 [256]byte = [256]byte{107, 114, 103, 121, 98, 109, 99, 119, 75, 82, 71, 89, 66, 77, 67, 87, 107, 98, 98, 98, 98, 66, 103, 99, 99, 99, 99, 66, 103, 99, 99, 99, 99, 66, 103, 99, 99, 99, 99, 66, 103, 99, 99, 99, 99, 66, 71, 71, 71, 71, 71, 67, 114, 109, 109, 109, 109, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 87, 87, 87, 87, 66, 71, 71, 71, 71, 71, 67, 114, 109, 109, 109, 109, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 87, 87, 87, 87, 66, 71, 71, 71, 71, 71, 67, 114, 109, 109, 109, 109, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 119, 119, 119, 87, 66, 121, 87, 87, 87, 87, 66, 71, 71, 71, 71, 71, 67, 114, 109, 109, 109, 109, 66, 121, 87, 87, 87, 87, 66, 121, 87, 87, 87, 87, 66, 121, 87, 87, 87, 87, 66, 121, 87, 87, 87, 87, 87, 71, 71, 71, 71, 87, 87, 82, 82, 82, 82, 82, 77, 82, 82, 82, 82, 82, 77, 82, 82, 82, 82, 82, 77, 82, 82, 82, 82, 82, 77, 82, 82, 82, 82, 87, 87, 89, 89, 89, 89, 87, 87, 107, 107, 107, 107, 75, 75, 75, 75, 119, 119, 119, 119, 119, 119, 119, 119, 87, 87, 87, 87, 87, 87, 87}

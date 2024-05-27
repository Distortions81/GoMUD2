package main

const ANSI_ESC = "\033["
const ANSI_RESET = ANSI_ESC + "m"
const NEWLINE = "\r\n"

var colorTable map[byte]*ctData = map[byte]*ctData{
	//bg colors
	'0': {code: "40", isBG: true}, //black
	'1': {code: "41", isBG: true}, //red
	'2': {code: "42", isBG: true}, //green
	'3': {code: "43", isBG: true}, //yellow
	'4': {code: "44", isBG: true}, //blue
	'5': {code: "45", isBG: true}, //magenta
	'6': {code: "46", isBG: true}, //cyan
	'7': {code: "47", isBG: true}, //white

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

	'!': {code: "1", disCode: "22", isStyle: true, style: bold},
	'*': {code: "3", disCode: "23", isStyle: true, style: italic},
	'_': {code: "4", disCode: "24", isStyle: true, style: underline},
	'^': {code: "7", disCode: "27", isStyle: true, style: inverse},
	'~': {code: "9", disCode: "29", isStyle: true, style: strike},
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
					out = append(out, []byte("\r\n")...)
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

// Combines multiple color codes, allows styles to be toggled on and off and ignores any code that would set/unset a state that is already set/unset
func ANSIColor(in []byte) []byte {
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

	//Process ANSI color/style
	out = []byte{}
	inLen = len(in)
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
				} else if in[x] == 'x' && s.hasColor {
					s.resetState()

					out = append(out, []byte(ANSI_RESET)...)
					continue
				}

				//Look up color/style
				cVal := colorTable[in[x]]

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
				//If new BH color, set it.
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
					cOut = append(cOut, colorTable['!'].code...)
				} else if !s.nextStyle.hasFlag(bold) && s.curStyle.hasFlag(bold) {
					//Remove bold style
					cOut = append(cOut, colorTable['!'].disCode...)
				} else if s.nextStyle.hasFlag(italic) && !s.curStyle.hasFlag(italic) {
					//Add italic style
					cOut = append(cOut, colorTable['*'].code...)
				} else if !s.nextStyle.hasFlag(italic) && s.curStyle.hasFlag(italic) {
					//Remove italic style
					cOut = append(cOut, colorTable['*'].disCode...)
				} else if s.nextStyle.hasFlag(underline) && !s.curStyle.hasFlag(underline) {
					//Add underline style
					cOut = append(cOut, colorTable['_'].code...)
				} else if !s.nextStyle.hasFlag(underline) && s.curStyle.hasFlag(underline) {
					//Remove underline style
					cOut = append(cOut, colorTable['_'].disCode...)
				} else if s.nextStyle.hasFlag(inverse) && !s.curStyle.hasFlag(inverse) {
					//Add inverse style
					cOut = append(cOut, colorTable['^'].code...)
				} else if !s.nextStyle.hasFlag(inverse) && s.curStyle.hasFlag(inverse) {
					//Remove inverse style
					cOut = append(cOut, colorTable['^'].disCode...)
				} else if s.nextStyle.hasFlag(strike) && !s.curStyle.hasFlag(strike) {
					//Add strike style
					cOut = append(cOut, colorTable['~'].code...)
				} else if !s.nextStyle.hasFlag(strike) && s.curStyle.hasFlag(strike) {
					//Remove strike style
					cOut = append(cOut, colorTable['~'].disCode...)
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
	bold = 1 << iota
	italic
	underline
	inverse
	strike
)

type ctData struct {
	code, disCode string
	style         Bitmask

	isBG, isFG, removeBold,
	isStyle bool
}

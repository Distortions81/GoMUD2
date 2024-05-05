package main

type Bitmask uint32

const ANSI_ESC = "\033["

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

func (f Bitmask) HasFlag(flag Bitmask) bool { return f&flag != 0 }
func (f *Bitmask) AddFlag(flag Bitmask)     { *f |= flag }
func (f *Bitmask) ClearFlag(flag Bitmask)   { *f &= ^flag }
func (f *Bitmask) ToggleFlag(flag Bitmask)  { *f ^= flag }

// Combines multiple color codes, allows styles to be toggled on and off and ignores any code that would set/unset a state that is already set/unset
func ANSIColor(i []byte) []byte {
	var (
		curStyle, nextStyle Bitmask
		curColor,
		curBGColor,

		nextColor,
		nextBGColor string
	)

	var out []byte
	il := len(i)

	for x := 0; x < il; x++ {
		//Color code
		if i[x] == '{' {
			x++
			if x < il {
				// escaped {{
				if i[x] == '{' {
					out = append(out, '{')
					continue
				} else if i[x] == 'x' {
					out = append(out, []byte("\033[0m")...)
					continue
				} else if i[x] == 'n' {
					out = append(out, []byte("\r\n")...)
					continue
				}
				val := colorTable[i[x]]
				if val == nil {
					continue
				}
				if val.isFG && curColor != val.code {
					nextColor = val.code
				}
				if val.isBG && curBGColor != val.code {
					nextBGColor = val.code
				}
				if !val.isFG {
					nextStyle.ToggleFlag(val.style)
				} else {
					nextStyle.AddFlag(val.style)
				}
				if val.removeBold {
					nextStyle.ClearFlag(bold)
				}
				continue
			} else {
				break
			}
		} else {

			if nextColor != "" || nextBGColor != "" || nextStyle != curStyle {
				var cout []byte
				if nextStyle.HasFlag(bold) && !curStyle.HasFlag(bold) {
					cout = append(cout, colorTable['!'].code...)
				} else if !nextStyle.HasFlag(bold) && curStyle.HasFlag(bold) {
					cout = append(cout, colorTable['!'].disCode...)
				}

				if nextStyle.HasFlag(italic) && !curStyle.HasFlag(italic) {
					cout = append(cout, colorTable['*'].code...)
				} else if !nextStyle.HasFlag(italic) && curStyle.HasFlag(italic) {
					cout = append(cout, colorTable['*'].disCode...)
				}

				if nextStyle.HasFlag(underline) && !curStyle.HasFlag(underline) {
					cout = append(cout, colorTable['_'].code...)
				} else if !nextStyle.HasFlag(underline) && curStyle.HasFlag(underline) {
					cout = append(cout, colorTable['_'].disCode...)
				}

				if nextStyle.HasFlag(inverse) && !curStyle.HasFlag(inverse) {
					cout = append(cout, colorTable['^'].code...)
				} else if !nextStyle.HasFlag(inverse) && curStyle.HasFlag(inverse) {
					cout = append(cout, colorTable['^'].disCode...)
				}

				if nextStyle.HasFlag(strike) && !curStyle.HasFlag(strike) {
					cout = append(cout, colorTable['~'].code...)
				} else if !nextStyle.HasFlag(strike) && curStyle.HasFlag(strike) {
					cout = append(cout, colorTable['~'].disCode...)
				}
				if nextBGColor != "" {
					if len(cout) > 0 {
						cout = append(cout, ';')
					}
					cout = append(cout, []byte(nextBGColor)...)
					nextBGColor = ""
				}
				if nextColor != "" {
					if nextColor != "" {
						cout = append(cout, ';')
					}
					cout = append(cout, []byte(nextColor)...)
					nextColor = ""
				}
				if len(cout) > 0 {
					cout = append(cout, 'm')

					curStyle = nextStyle
					curColor = nextColor
					curBGColor = nextBGColor

					out = append(out, []byte(ANSI_ESC)...)
					out = append(out, cout...)
				}
			}
		}
		out = append(out, i[x])
	}
	return append(out, []byte("\033[0m")...)
}

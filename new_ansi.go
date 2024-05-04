package main

type Bitmask uint32

const (
	bold = 1 << iota
	italic
	underline
	inverse
	strike
)

var (
	curStyle, nextStyle Bitmask
	curColor,
	curBGColor,

	nextColor,
	nextBGColor string
)

type ctData struct {
	code, disCode string
	style         Bitmask

	isBG, isFG, notBold,
	isStyle bool
}

var colorTable map[byte]*ctData = map[byte]*ctData{
	'0': {code: "40", isBG: true},
	'1': {code: "41", isBG: true},
	'2': {code: "42", isBG: true},
	'3': {code: "43", isBG: true},
	'4': {code: "44", isBG: true},
	'5': {code: "45", isBG: true},
	'6': {code: "46", isBG: true},
	'7': {code: "47", isBG: true},

	'k': {code: "30", isFG: true, notBold: true},
	'r': {code: "31", isFG: true, notBold: true},
	'g': {code: "32", isFG: true, notBold: true},
	'y': {code: "33", isFG: true, notBold: true},
	'b': {code: "34", isFG: true, notBold: true},
	'm': {code: "35", isFG: true, notBold: true},
	'c': {code: "36", isFG: true, notBold: true},
	'w': {code: "37", isFG: true, notBold: true},

	'K': {code: "30", isFG: true, style: bold},
	'R': {code: "31", isFG: true, style: bold},
	'G': {code: "32", isFG: true, style: bold},
	'Y': {code: "33", isFG: true, style: bold},
	'B': {code: "34", isFG: true, style: bold},
	'M': {code: "35", isFG: true, style: bold},
	'C': {code: "36", isFG: true, style: bold},
	'W': {code: "37", isFG: true, style: bold},

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

func ANSIColor(i []byte) []byte {
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
				if val.style != bold {
					nextStyle.ToggleFlag(val.style)
				} else {
					nextStyle.AddFlag(val.style)
				}
				if val.notBold {
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
				if nextColor != "" {
					if len(cout) > 0 {
						cout = append(cout, ';')
					}
					cout = append(cout, []byte(nextColor)...)
				}
				if nextBGColor != "" {
					if nextColor != "" {
						cout = append(cout, ';')
					}
					cout = append(cout, []byte(nextColor)...)
				}
				if len(cout) > 0 {
					cout = append(cout, 'm')

					curStyle = nextStyle
					curColor = nextColor
					curBGColor = nextBGColor
					nextColor = ""
					nextBGColor = ""

					out = append(out, []byte("\033[")...)
					out = append(out, cout...)
				}
			}
		}
		out = append(out, i[x])
	}
	return append(out, []byte("\033[0m")...)
}

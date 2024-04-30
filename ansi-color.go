package main

import "strings"

type cModes struct {
	reset, bright, italic, underline, blink, inverse, strike bool
}

// Subsitute color codes with ANSI color, with a color reset at the end if needed
func ANSIColor(in string) string {
	didColor := false

	output := ""
	input := in

	var length int
	var lfgColor, lbgColor string
	var cm cModes

	for i := 0; ; i++ {
		length = len(input)

		if i < length {
			cur := input[i]
			if i+1 < length {
				next := input[i+1]

				if cur == '{' {
					if next == '{' {
						output = input[:i] + "{" + input[i+2:]
						input = output
						continue
					}
					if next == 'n' {
						output = input[:i] + "\r\n" + input[i+2:]
						input = output
						continue
					}
					nm, fgcolor, bgcolor := convertColor(next)
					if nm.reset && !cm.reset {
						fgcolor = ""
						bgcolor = ""
						lfgColor = ""
						lbgColor = ""
						cm.blink = false
						cm.bright = false
						cm.inverse = false
						cm.italic = false
						cm.strike = false
						cm.underline = false
						didColor = false
					}
					if fgcolor != "" {
						if strings.EqualFold(fgcolor, lfgColor) {
							output = input[:i] + input[i+2:]
							input = output
							continue
						} else {
							lfgColor = fgcolor
						}
					}

					if bgcolor != "" {
						if bgcolor == lbgColor {
							output = input[:i] + input[i+2:]
							input = output
							continue
						} else {
							lbgColor = bgcolor
							fgcolor = bgcolor
						}
					}

					am := cModes{}
					dm := cModes{}

					if nm.reset {
						am.reset = true
					}
					cm.reset = nm.reset

					if !cm.bright && nm.bright {
						am.bright = true
						cm.bright = true
					} else if cm.bright && !nm.bright {
						dm.bright = true
						cm.bright = false
					}

					if !cm.italic && nm.italic {
						am.italic = true
						cm.italic = true
					} else if cm.italic && nm.italic {
						dm.italic = true
						cm.italic = false
					}

					if !cm.underline && nm.underline {
						am.underline = true
						cm.underline = true
					} else if cm.underline && nm.underline {
						dm.underline = true
						cm.underline = false
					}

					if !cm.blink && nm.blink {
						am.blink = true
						cm.blink = true
					} else if cm.blink && nm.blink {
						dm.blink = true
						cm.blink = false
					}

					if !cm.inverse && nm.inverse {
						am.inverse = true
						cm.inverse = true
					} else if cm.inverse && nm.inverse {
						dm.inverse = true
						cm.inverse = false
					}

					if !cm.strike && nm.strike {
						am.strike = true
						cm.strike = true
					} else if cm.strike && nm.strike {
						dm.strike = true
						cm.strike = false
					}

					//TODO: Combine successive codes
					ansiCode := getColorNew(am, dm, fgcolor)

					//Remove color code, and insert ANSI sequence
					if !am.reset {
						didColor = true
					}
					output = input[:i] + ansiCode + input[i+2:]
					input = output

				}
			}
		} else {
			break
		}
	}
	if didColor {
		input = input + getColorNew(cModes{reset: true}, cModes{}, "")
	}
	return input

}

// Just strip color codes and produce normal text
func StripColorCodes(in string) string {
	output := ""
	input := in

	for i := 0; ; i++ {
		length := len(input) - 1
		if i < length {
			cur := input[i]

			if i+1 < length {
				next := input[i+1]

				if cur == '{' {
					if next == '{' {
						output = input[:i] + "{" + input[i+2:]
						input = output
					} else if next == 'n' {
						output = input[:i] + "\r\n" + input[i+2:]
						input = output
						continue
					} else {
						output = input[:i] + input[i+2:]
						input = output
					}
				}
			}
		} else {
			break
		}
	}
	return input

}

// Convert a single color code character to a ANSI color sequence
func convertColor(i byte) (modes cModes, fgColor, bgColor string) {
	switch i {

	//Styles
	case 'x': //reset
		return cModes{reset: true}, "", ""
	case 'l': //bright
		return cModes{bright: true}, "", ""
	case 'i': //italic
		return cModes{italic: true}, "", ""
	case 'u': //underline
		return cModes{underline: true}, "", ""
	case '=': //blink
		return cModes{blink: true}, "", ""
	case 'v': //inverse
		return cModes{inverse: true}, "", ""
	case 's': //strikethrough
		return cModes{strike: true}, "", ""

	//Foreground colors
	case 'k': //black
		return cModes{}, "30", ""
	case 'r': //red
		return cModes{}, "31", ""
	case 'g': //green
		return cModes{}, "32", ""
	case 'y': //yellow
		return cModes{}, "33", ""
	case 'b': //blue
		return cModes{}, "34", ""
	case 'm': //magenta
		return cModes{}, "35", ""
	case 'c': //cyan
		return cModes{}, "36", ""
	case 'w': //white
		return cModes{}, "37", ""
	case 'd': //default
		return cModes{}, "39", ""

	//Light foreground colors
	case 'K': //black
		return cModes{bright: true}, "30", ""
	case 'R': //red
		return cModes{bright: true}, "31", ""
	case 'G': //green
		return cModes{bright: true}, "32", ""
	case 'Y': //yellow
		return cModes{bright: true}, "33", ""
	case 'B': //blue
		return cModes{bright: true}, "34", ""
	case 'M': //magenta
		return cModes{bright: true}, "35", ""
	case 'C': //cyan
		return cModes{bright: true}, "36", ""
	case 'W': //white
		return cModes{bright: true}, "37", ""

	//Background colors
	case '!': //black
		return cModes{}, "", "40"
	case '@': //red
		return cModes{}, "", "41"
	case '#': //green
		return cModes{}, "", "42"
	case '$': //yellow
		return cModes{}, "", "43"
	case '%': //blue
		return cModes{}, "", "44"
	case '^': //magenta
		return cModes{}, "", "45"
	case '&': //cyan
		return cModes{}, "", "46"
	case '*': //white
		return cModes{}, "", "47"
	case 'D': //default
		return cModes{}, "", "49"

	default:
		return cModes{}, "", ""
	}
}

func getColorNew(add, del cModes, color string) string {
	buf := "\033["
	var numArgs int

	if add.reset || del.reset {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "0"
		numArgs++
	}
	if add.bright {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "1"
		numArgs++
	}
	if add.italic {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "3"
		numArgs++
	}
	if add.underline {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "4"
		numArgs++
	}
	if add.blink {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "5"
		numArgs++
	}
	if add.inverse {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "7"
		numArgs++
	}
	if add.strike {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "9"
		numArgs++
	}

	//Remove mode
	if del.bright {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "22"
		numArgs++
	}
	if del.italic {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "23"
		numArgs++
	}
	if del.underline {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "24"
		numArgs++
	}
	if del.blink {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "25"
		numArgs++
	}
	if del.inverse {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "27"
		numArgs++
	}
	if del.strike {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + "29"
		numArgs++
	}

	if color != "" {
		if numArgs > 0 {
			buf = buf + ";"
		}
		buf = buf + color
		numArgs++
	}

	if numArgs > 0 {
		buf = buf + "m"
		return buf
	} else {
		return ""
	}
}

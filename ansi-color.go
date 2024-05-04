package main

func aNSIColor(input string) string {
	output := ""
	length := len(input)
	var lastColor string
	var wasBold bool
	for i := 0; i < length; i++ {
		cur := input[i]
		var next byte
		if i+1 < length {
			next = input[i+1]
		}

		if cur == '{' {
			if next == '{' {
				continue
			} else if next == 'x' {
				if lastColor != "" || wasBold {
					output = output + ESC + "0m"
					wasBold = false
					lastColor = ""
				}
				i++
				continue
			} else {
				color, bold := codeToANSI(next)
				if color != "" && (color != lastColor || bold != wasBold) {
					if bold != wasBold {
						if bold {
							output = output + ESC + "1;" + color + "m"
						} else {
							output = output + ESC + "0;" + color + "m"
						}
						wasBold = bold
					} else {
						output = output + ESC + color + "m"
					}

					if color != "" {
						lastColor = color
					}
				}

				i++
				continue
			}
		} else {
			output = output + string(cur)
		}
	}
	if lastColor != "" || wasBold {
		output = output + ESC + "0m"
	}
	return output
}

// Just strip color codes and produce normal text
func StripColorCodes(input string) string {
	output := ""
	length := len(input)
	for i := 0; i < length; i++ {
		cur := input[i]
		var next byte
		if i+1 < length {
			next = input[i+1]
		}

		if cur == '{' {
			if next == '{' {
				continue
			} else {
				i++
				continue
			}
		} else {
			output = output + string(cur)
		}
	}
	return output
}

func codeToANSI(color byte) (string, bool) {

	switch color {
	case 'k':
		return C_BLACK, false
	case 'r':
		return C_RED, false
	case 'g':
		return C_GREEN, false
	case 'y':
		return C_YELLOW, false
	case 'b':
		return C_BLUE, false
	case 'm':
		return C_MAGENTA, false
	case 'c':
		return C_CYAN, false
	case 'w':
		return C_WHITE, false

	case 'K':
		return C_BLACK, true
	case 'R':
		return C_RED, true
	case 'G':
		return C_GREEN, true
	case 'Y':
		return C_YELLOW, true
	case 'B':
		return C_BLUE, true
	case 'M':
		return C_MAGENTA, true
	case 'C':
		return C_CYAN, true
	case 'W':
		return C_WHITE, true
	default:
		return "", false
	}
}

const (
	//Octal codes \000
	BELL      = "\007"
	BACKSPACE = "\010"
	HTAB      = "\011"
	VTAB      = "\013"
	FORMFEED  = "\014"
	NEWLINE   = "\015\012"
	ESC       = "\033["
	DEL_CHAR  = "\0177"
)

// Enable style
const (
	C_NONE    = "0"
	C_BOLD    = "1"
	C_DIM     = "2"
	C_ITALIC  = "3"
	C_UNDER   = "4"
	C_INVERSE = "7"
	C_STRIKE  = "9"
)

// Disable style
const (
	R_BOLD    = "22"
	R_DIM     = "22"
	R_ITALIC  = "23"
	R_UNDER   = "24"
	R_INVERSE = "27"
	R_STRIKE  = "29"
)

// Foreground color
const (
	C_BLACK   = "30"
	C_RED     = "31"
	C_GREEN   = "32"
	C_YELLOW  = "33"
	C_BLUE    = "34"
	C_MAGENTA = "35"
	C_CYAN    = "36"
	C_WHITE   = "37"
	C_DEFAULT = "39"
)

// Foreground color
const (
	CB_BLACK   = "30"
	CB_RED     = "31"
	CB_GREEN   = "32"
	CB_YELLOW  = "33"
	CB_BLUE    = "34"
	CB_MAGENTA = "35"
	CB_CYAN    = "36"
	CB_WHITE   = "37"
	CB_DEFAULT = "39"
)

// Background color
const (
	BG_BLACK   = "40"
	BG_RED     = "41"
	BG_GREEN   = "42"
	BG_YELLOW  = "43"
	BG_BLUE    = "44"
	BG_MAGENTA = "45"
	BG_CYAN    = "46"
	BG_WHITE   = "47"
	BG_DEFAULT = "49"
)

// aixterm bright colors
const (
	XB_BLACK   = "90"
	XB_RED     = "91"
	XB_GREEN   = "92"
	XB_YELLOW  = "93"
	XB_BLUE    = "94"
	XB_MAGENTA = "95"
	XB_CYAN    = "96"
	XB_WHITE   = "97"
)

// aixterm bright background color
const (
	XG_BLACK   = "90"
	XG_RED     = "91"
	XG_GREEN   = "92"
	XG_YELLOW  = "93"
	XG_BLUE    = "94"
	XG_MAGENTA = "95"
	XG_CYAN    = "96"
	XG_WHITE   = "97"
)

// 8-bit color
const (
	EB_FG = "38;5{"
	EB_BG = "48;5"
)

// 24-bit color
const (
	TFB_FG = "38;2"
	TFB_BG = "48;2"
)

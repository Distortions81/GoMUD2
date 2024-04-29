package main

import "strings"

func txtTo7bit(data string) string {
	var tmp string
	// Filter to 7-bit
	for _, data := range data {
		if (data >= '0' && data <= '9') ||
			(data >= 'A' && data <= 'Z') ||
			(data >= 'a' && data <= 'z') {
			tmp = tmp + string(data)
		}
	}

	return tmp
}

func telSnFilter(data string) string {
	ttype := txtTo7bit(data)
	ttype = strings.ToUpper(ttype)
	return ttype
}

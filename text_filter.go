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

// Returns false if name is prohibited
func nameReserved(name string) bool {
	for _, item := range reservedNames {
		if item == name {
			return true
		}
	}

	return false
}

func nameReduce(name string) string {
	name = strings.ToLower(name)
	var newName string
	for _, l := range name {
		if l >= 'a' && l <= 'z' {
			newName = newName + string(l)
		}
	}
	return toTitle(newName)
}

func toTitle(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func fileSafeName(name string) string {
	var newName string
	for _, l := range name {
		if l >= 'a' && l <= 'z' ||
			l >= 'A' && l <= 'Z' ||
			l >= '0' && l <= '9' {
			newName = newName + string(l)
		} else if l == ' ' {
			newName = newName + "_"
		}
	}
	return toTitle(newName)
}

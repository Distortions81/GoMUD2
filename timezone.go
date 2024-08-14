package main

import (
	"os"
	"strings"
	"unicode"
)

var tzRegions map[string][]string

func GetOsTimeZones() {
	zones := []string{}
	var zoneDirs = []string{
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}

	for _, zd := range zoneDirs {
		zones = walkTzDir(zd, zones)

		for idx, zone := range zones {
			zones[idx] = strings.ReplaceAll(zone, zd+"/", "")
		}
	}

	tzRegions = make(map[string][]string)
	for _, item := range zones {
		args := strings.Split(item, "/")
		if len(args) >= 2 {
			tzRegions[args[0]] = append(tzRegions[args[0]], args[1])
		}
	}
}

func walkTzDir(path string, zones []string) []string {
	fileInfos, err := os.ReadDir(path)
	if err != nil {
		return zones
	}

	isAlpha := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}

	for _, info := range fileInfos {
		if info.Name() != strings.ToUpper(info.Name()[:1])+info.Name()[1:] {
			continue
		}

		if !isAlpha(info.Name()[:1]) {
			continue
		}

		newPath := path + "/" + info.Name()

		if info.IsDir() {
			zones = walkTzDir(newPath, zones)
		} else {
			zones = append(zones, newPath)
		}
	}

	return zones
}

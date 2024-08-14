package main

import "time"

type noteTypeData struct {
	Version int
	UUID    uuidData

	Name             string
	File             string
	PostLvl, ReadLvl int
	Notes            []*noteData
	Modified         time.Time

	Disabled bool
	dirty    bool
}

type noteWhoData struct {
	Name string   `json:",omitempty"`
	UUID uuidData `json:",omitempty"`
	//guildid
}

type noteData struct {
	From noteWhoData

	WordWrap bool
	To       []noteWhoData
	CC       []noteWhoData `json:",omitempty"`
	BCC      []noteWhoData `json:",omitempty"`
	Subject  string
	Text     string

	Hidden bool

	Created  time.Time
	Modified time.Time
}

var noteTypes []noteTypeData
